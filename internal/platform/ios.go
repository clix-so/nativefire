package platform

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/clix-so/nativefire/internal/firebase"
	"github.com/clix-so/nativefire/internal/ui"
)

// Constants for repeated strings
const (
	googleServiceInfoPlist = "GoogleService-Info.plist"
	iosName                = "iOS"
)

func (p *IOSPlatform) Name() string {
	return iosName
}

func (p *IOSPlatform) Type() Type {
	return iOS
}

func (p *IOSPlatform) Detect() bool {
	return fileExists(iosString) ||
		findFile(".", "*.xcodeproj") != "" ||
		findFile(".", "*.xcworkspace") != "" ||
		fileExists("Podfile")
}

func (p *IOSPlatform) ConfigFileName() string {
	return googleServiceInfoPlist
}

func (p *IOSPlatform) ConfigPath() string {
	if fileExists(iosString) {
		return iosString
	}
	return "."
}

func (p *IOSPlatform) InstallConfig(config *firebase.Config) error {
	configPath := p.ConfigPath()

	runnerPath := filepath.Join(configPath, "Runner")
	if fileExists(runnerPath) {
		configPath = runnerPath
	}

	projectName := p.findProjectName()
	if projectName != "" {
		projectPath := filepath.Join(configPath, projectName)
		if fileExists(projectPath) {
			configPath = projectPath
		}
	}

	targetPath := filepath.Join(configPath, p.ConfigFileName())

	if err := os.MkdirAll(configPath, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", configPath, err)
	}

	// Use the unique temp file path from config instead of hardcoded temp location
	sourceFile := config.ConfigFile
	if sourceFile == "" {
		// Fallback to old behavior if ConfigFile is not set
		sourceFile = filepath.Join(os.TempDir(), p.ConfigFileName())
	}

	sourceData, err := os.ReadFile(sourceFile)
	if err != nil {
		return fmt.Errorf("failed to read source config file: %w", err)
	}

	if err := os.WriteFile(targetPath, sourceData, 0644); err != nil {
		ui.AnimatedError("Failed to write configuration file")
		ui.WarningMsg("Please add GoogleService-Info.plist to your Xcode project manually")
		return fmt.Errorf("failed to write config file to %s: %w", targetPath, err)
	}

	// Clean up the temp file after successful installation
	if config.ConfigFile != "" {
		os.Remove(config.ConfigFile)
	}

	ui.AnimatedSuccess(fmt.Sprintf("Configuration file installed at: %s", targetPath))
	return nil
}

func (p *IOSPlatform) AddInitializationCode(config *firebase.Config) error {
	podfilePath := p.findPodfile()
	podsAdded := false
	spmSetup := false

	if podfilePath != "" {
		if err := p.addFirebasePods(podfilePath); err != nil {
			return err
		}
		podsAdded = true
	} else if p.hasSwiftPackages() {
		// Handle existing Package.swift projects
		if err := p.setupSPMPackageSwift(); err != nil {
			return nil // User chose to skip
		}
		spmSetup = true
	} else if p.shouldUseSPM() {
		// Handle Xcode projects that should use SPM
		if err := p.setupSPMFirebase(); err != nil {
			return nil // User chose to skip
		}
		spmSetup = true
	}

	appDelegatePath := p.findAppDelegate()
	if appDelegatePath == "" {
		// AppDelegate not found, create one with animation
		var newAppDelegatePath string
		err := ui.ShowLoader("Creating AppDelegate", func() error {
			var createErr error
			newAppDelegatePath, createErr = p.createAppDelegate()
			return createErr
		})

		if err != nil {
			ui.AnimatedError("Failed to create AppDelegate")
			ui.WarningMsg("Please manually add Firebase initialization code")
			return nil
		}
		appDelegatePath = newAppDelegatePath
		ui.AnimatedSuccess(fmt.Sprintf("Created AppDelegate at: %s", appDelegatePath))
	}

	// Add Firebase initialization with loading animation
	err := ui.ShowLoader("Adding Firebase initialization code", func() error {
		return p.addFirebaseInitialization(appDelegatePath)
	})
	if err != nil {
		return err
	}

	// Run package manager commands after all changes
	if podsAdded || spmSetup {
		return p.runPackageManagerCommands()
	}

	return nil
}

func (p *IOSPlatform) findProjectName() string {
	xcodeproj := findFile(".", "*.xcodeproj")
	if xcodeproj != "" {
		return strings.TrimSuffix(filepath.Base(xcodeproj), ".xcodeproj")
	}
	return ""
}

func (p *IOSPlatform) findPodfile() string {
	candidates := []string{
		"Podfile",
		"ios/Podfile",
	}

	for _, candidate := range candidates {
		if fileExists(candidate) {
			return candidate
		}
	}
	return ""
}

func (p *IOSPlatform) findAppDelegate() string {
	appDelegatePath := findFile(".", "AppDelegate.swift")
	if appDelegatePath == "" {
		appDelegatePath = findFile(".", "AppDelegate.m")
	}
	return appDelegatePath
}

func (p *IOSPlatform) addFirebasePods(podfilePath string) error {
	content, err := os.ReadFile(podfilePath)
	if err != nil {
		return fmt.Errorf("failed to read Podfile: %w", err)
	}

	contentStr := string(content)

	if !strings.Contains(contentStr, "Firebase/Core") {
		lines := strings.Split(contentStr, "\n")
		var newLines []string

		for _, line := range lines {
			newLines = append(newLines, line)
			if strings.Contains(line, "target") && strings.Contains(line, "do") {
				newLines = append(newLines, "  pod 'Firebase/Core'")
				newLines = append(newLines, "  pod 'Firebase/Analytics'")
			}
		}

		newContent := strings.Join(newLines, "\n")
		if err := os.WriteFile(podfilePath, []byte(newContent), 0644); err != nil {
			return fmt.Errorf("failed to update Podfile: %w", err)
		}

		ui.SuccessMsg(fmt.Sprintf("Added Firebase pods to: %s", podfilePath))
	}

	return nil
}

func (p *IOSPlatform) checkFirebaseAppDelegateProxy() (bool, error) {
	// Check for FirebaseAppDelegateProxyEnabled in Info.plist files
	infoPlistFiles := []string{
		"ios/Runner/Info.plist",
		"Info.plist",
		"Runner/Info.plist",
	}

	for _, file := range infoPlistFiles {
		if !fileExists(file) {
			continue
		}

		content, err := os.ReadFile(file)
		if err != nil {
			continue
		}

		contentStr := string(content)
		if strings.Contains(contentStr, "FirebaseAppDelegateProxyEnabled") {
			// Look for the value after FirebaseAppDelegateProxyEnabled
			lines := strings.Split(contentStr, "\n")
			for i, line := range lines {
				if strings.Contains(line, "FirebaseAppDelegateProxyEnabled") && i+1 < len(lines) {
					nextLine := strings.TrimSpace(lines[i+1])
					if strings.Contains(nextLine, "<false/>") || strings.Contains(nextLine, "<false></false>") {
						return false, nil // Proxy is disabled
					}
				}
			}
		}
	}

	// Default is true (proxy enabled) if not specified
	return true, nil
}

func (p *IOSPlatform) addFirebaseInitialization(appDelegatePath string) error {
	// Check if FirebaseAppDelegateProxyEnabled is disabled
	proxyEnabled, err := p.checkFirebaseAppDelegateProxy()
	if err != nil {
		ui.WarningMsg("Could not check FirebaseAppDelegateProxyEnabled setting, proceeding with default configuration")
	}

	content, err := os.ReadFile(appDelegatePath)
	if err != nil {
		return fmt.Errorf("failed to read AppDelegate: %w", err)
	}

	contentStr := string(content)

	if strings.Contains(appDelegatePath, ".swift") {
		return p.addSwiftFirebaseInitialization(contentStr, appDelegatePath, proxyEnabled)
	} else if strings.Contains(appDelegatePath, ".m") {
		return p.addObjCFirebaseInitialization(contentStr, appDelegatePath, proxyEnabled)
	}

	return fmt.Errorf("unsupported AppDelegate file type: %s", appDelegatePath)
}

func (p *IOSPlatform) addSwiftFirebaseInitialization(contentStr, appDelegatePath string, proxyEnabled bool) error {
	// Check if Firebase is already configured
	if strings.Contains(contentStr, "FirebaseApp.configure()") {
		ui.InfoMsg("Firebase is already configured in AppDelegate")
		return nil
	}

	// Add FirebaseCore import if not present
	if !strings.Contains(contentStr, "import FirebaseCore") {
		// Replace import Firebase with import FirebaseCore if it exists
		if strings.Contains(contentStr, "import Firebase") {
			contentStr = strings.Replace(contentStr, "import Firebase", "import FirebaseCore", 1)
		} else {
			// Add new import
			if strings.Contains(contentStr, "import UIKit") {
				contentStr = strings.Replace(contentStr,
					"import UIKit",
					"import UIKit\nimport FirebaseCore", 1)
			} else {
				// Add import at the beginning
				contentStr = "import FirebaseCore\n" + contentStr
			}
		}
	}

	// Add FirebaseApp.configure() in didFinishLaunchingWithOptions
	if strings.Contains(contentStr, "didFinishLaunchingWithOptions") {
		const swiftMethod = "func application(_ application: UIApplication, " +
			"didFinishLaunchingWithOptions launchOptions: [UIApplication.LaunchOptionsKey: Any]?) -> Bool {"
		const swiftMethodWithFirebase = "func application(_ application: UIApplication, " +
			"didFinishLaunchingWithOptions launchOptions: [UIApplication.LaunchOptionsKey: Any]?) -> Bool {" +
			"\n        FirebaseApp.configure()"
		contentStr = strings.Replace(contentStr, swiftMethod, swiftMethodWithFirebase, 1)
	}

	// If FirebaseAppDelegateProxyEnabled is disabled, add required delegate methods
	if !proxyEnabled {
		contentStr = p.addSwiftDelegateMethods(contentStr)
		ui.InfoMsg("Added Firebase delegate methods (FirebaseAppDelegateProxyEnabled is disabled)")
	}

	if err := os.WriteFile(appDelegatePath, []byte(contentStr), 0644); err != nil {
		return fmt.Errorf("failed to update AppDelegate: %w", err)
	}

	ui.SuccessMsg(fmt.Sprintf("Added Firebase initialization to: %s", appDelegatePath))
	if !proxyEnabled {
		ui.InfoMsg("Added additional delegate methods for manual Firebase integration")
	}
	return nil
}

func (p *IOSPlatform) addObjCFirebaseInitialization(contentStr, appDelegatePath string, proxyEnabled bool) error {
	// Check if Firebase is already configured
	if strings.Contains(contentStr, "[FIRApp configure]") {
		ui.InfoMsg("Firebase is already configured in AppDelegate")
		return nil
	}

	// Add FirebaseCore import if not present
	if !strings.Contains(contentStr, "@import FirebaseCore;") && !strings.Contains(contentStr, "#import <FirebaseCore/FirebaseCore.h>") {
		// Replace existing Firebase imports with FirebaseCore
		if strings.Contains(contentStr, "@import Firebase;") {
			contentStr = strings.Replace(contentStr, "@import Firebase;", "@import FirebaseCore;", 1)
		} else if strings.Contains(contentStr, "#import <Firebase/Firebase.h>") {
			contentStr = strings.Replace(contentStr, "#import <Firebase/Firebase.h>", "#import <FirebaseCore/FirebaseCore.h>", 1)
		} else {
			// Add new import
			if strings.Contains(contentStr, "#import \"AppDelegate.h\"") {
				contentStr = strings.Replace(contentStr,
					"#import \"AppDelegate.h\"",
					"#import \"AppDelegate.h\"\n@import FirebaseCore;", 1)
			} else {
				// Add import at the beginning
				contentStr = "@import FirebaseCore;\n" + contentStr
			}
		}
	}

	// Add [FIRApp configure]; in didFinishLaunchingWithOptions
	if strings.Contains(contentStr, "didFinishLaunchingWithOptions") {
		const objcMethod = "- (BOOL)application:(UIApplication *)application " +
			"didFinishLaunchingWithOptions:(NSDictionary *)launchOptions {"
		const objcMethodWithFirebase = "- (BOOL)application:(UIApplication *)application " +
			"didFinishLaunchingWithOptions:(NSDictionary *)launchOptions {\n    [FIRApp configure];"
		contentStr = strings.Replace(contentStr, objcMethod, objcMethodWithFirebase, 1)
	}

	// If FirebaseAppDelegateProxyEnabled is disabled, add required delegate methods
	if !proxyEnabled {
		contentStr = p.addObjCDelegateMethods(contentStr)
		ui.InfoMsg("Added Firebase delegate methods (FirebaseAppDelegateProxyEnabled is disabled)")
	}

	if err := os.WriteFile(appDelegatePath, []byte(contentStr), 0644); err != nil {
		return fmt.Errorf("failed to update AppDelegate: %w", err)
	}

	ui.SuccessMsg(fmt.Sprintf("Added Firebase initialization to: %s", appDelegatePath))
	if !proxyEnabled {
		ui.InfoMsg("Added additional delegate methods for manual Firebase integration")
	}
	return nil
}

func (p *IOSPlatform) addSwiftDelegateMethods(contentStr string) string {
	// Add required delegate methods for push notifications when FirebaseAppDelegateProxyEnabled is NO
	delegateMethods := `
    // MARK: - Firebase Push Notification Delegate Methods
    func application(_ application: UIApplication, didRegisterForRemoteNotificationsWithDeviceToken deviceToken: Data) {
        Messaging.messaging().apnsToken = deviceToken
    }
    
    func application(_ application: UIApplication, didFailToRegisterForRemoteNotificationsWithError error: Error) {
        print("Failed to register for remote notifications: \(error)")
    }
    
    func application(_ application: UIApplication, didReceiveRemoteNotification userInfo: [AnyHashable: Any]) {
        // Handle background notification
    }
    
    func application(_ application: UIApplication, 
                     didReceiveRemoteNotification userInfo: [AnyHashable: Any], 
                     fetchCompletionHandler completionHandler: @escaping (UIBackgroundFetchResult) -> Void) {
        // Handle background notification with completion handler
        completionHandler(.newData)
    }`

	// Add import for Firebase Messaging if not present
	if !strings.Contains(contentStr, "import FirebaseMessaging") {
		contentStr = strings.Replace(contentStr,
			"import Firebase",
			"import Firebase\nimport FirebaseMessaging", 1)
	}

	// Find the end of the class and add delegate methods before the closing brace
	if strings.Contains(contentStr, "@UIApplicationMain") || strings.Contains(contentStr, "class AppDelegate") {
		// Find the last closing brace of the class
		lastBraceIndex := strings.LastIndex(contentStr, "}")
		if lastBraceIndex != -1 {
			contentStr = contentStr[:lastBraceIndex] + delegateMethods + "\n" + contentStr[lastBraceIndex:]
		}
	}

	return contentStr
}

func (p *IOSPlatform) addObjCDelegateMethods(contentStr string) string {
	// Add required delegate methods for push notifications when FirebaseAppDelegateProxyEnabled is NO
	delegateMethods := `
#pragma mark - Firebase Push Notification Delegate Methods

- (void)application:(UIApplication *)application 
didRegisterForRemoteNotificationsWithDeviceToken:(NSData *)deviceToken {
    [FIRMessaging messaging].APNSToken = deviceToken;
}

- (void)application:(UIApplication *)application didFailToRegisterForRemoteNotificationsWithError:(NSError *)error {
    NSLog(@"Failed to register for remote notifications: %@", error);
}

- (void)application:(UIApplication *)application didReceiveRemoteNotification:(NSDictionary *)userInfo {
    // Handle background notification
}

- (void)application:(UIApplication *)application
                     didReceiveRemoteNotification:(NSDictionary *)userInfo
                           fetchCompletionHandler:(void (^)(UIBackgroundFetchResult))completionHandler {
    // Handle background notification with completion handler
    completionHandler(UIBackgroundFetchResultNewData);
}`

	// Add import for Firebase Messaging if not present
	if !strings.Contains(contentStr, "@import FirebaseMessaging;") &&
		!strings.Contains(contentStr, "#import <FirebaseMessaging/FirebaseMessaging.h>") {
		contentStr = strings.Replace(contentStr,
			"@import Firebase;",
			"@import Firebase;\n@import FirebaseMessaging;", 1)
	}

	// Find the end of the implementation and add delegate methods before @end
	if strings.Contains(contentStr, "@end") {
		endIndex := strings.LastIndex(contentStr, "@end")
		if endIndex != -1 {
			contentStr = contentStr[:endIndex] + delegateMethods + "\n\n" + contentStr[endIndex:]
		}
	}

	return contentStr
}

func (p *IOSPlatform) createAppDelegate() (string, error) {
	// Determine project language and structure
	isSwift := p.isSwiftProject()
	projectPath := p.determineAppDelegatePath()

	if isSwift {
		// Check if this is a SwiftUI project
		if p.isSwiftUIProject() {
			return p.createSwiftUIAppDelegateIntegration(projectPath)
		} else {
			return p.createSwiftAppDelegate(projectPath)
		}
	} else {
		return p.createObjCAppDelegate(projectPath)
	}
}

func (p *IOSPlatform) isSwiftProject() bool {
	// Check for existing Swift files
	if findFile(".", "*.swift") != "" {
		return true
	}

	// Check for Objective-C files
	if findFile(".", "*.m") != "" || findFile(".", "*.h") != "" {
		return false
	}

	// Check Podfile for Swift usage
	podfilePath := p.findPodfile()
	if podfilePath != "" {
		content, err := os.ReadFile(podfilePath)
		if err == nil {
			contentStr := string(content)
			if strings.Contains(contentStr, "use_frameworks!") {
				return true
			}
		}
	}

	// Default to Swift for new projects
	return true
}

func (p *IOSPlatform) determineAppDelegatePath() string {
	// Check for existing project structure
	projectName := p.findProjectName()

	// Try different common locations
	candidates := []string{
		"ios",
		".",
	}

	if projectName != "" {
		candidates = append([]string{
			filepath.Join("ios", projectName),
			projectName,
		}, candidates...)
	}

	// Use the first existing directory, or current directory as fallback
	for _, candidate := range candidates {
		if fileExists(candidate) {
			return candidate
		}
	}

	return "."
}

func (p *IOSPlatform) createSwiftAppDelegate(projectPath string) (string, error) {
	appDelegatePath := filepath.Join(projectPath, "AppDelegate.swift")

	// Check if file already exists
	if fileExists(appDelegatePath) {
		return appDelegatePath, nil
	}

	// Create directory if needed
	if err := os.MkdirAll(projectPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory %s: %w", projectPath, err)
	}

	swiftContent := `import UIKit
import FirebaseCore

@main
class AppDelegate: UIResponder, UIApplicationDelegate {

    var window: UIWindow?

    func application(_ application: UIApplication,
                     didFinishLaunchingWithOptions launchOptions: [UIApplication.LaunchOptionsKey: Any]?) -> Bool {
        // Configure Firebase
        FirebaseApp.configure()
        
        // Override point for customization after application launch.
        return true
    }

    // MARK: UISceneSession Lifecycle (iOS 13+)
    @available(iOS 13.0, *)
    func application(_ application: UIApplication,
                     configurationForConnecting connectingSceneSession: UISceneSession,
                     options: UIScene.ConnectionOptions) -> UISceneConfiguration {
        return UISceneConfiguration(name: "Default Configuration", sessionRole: connectingSceneSession.role)
    }

    @available(iOS 13.0, *)
    func application(_ application: UIApplication, didDiscardSceneSessions sceneSessions: Set<UISceneSession>) {
        // Called when the user discards a scene session.
    }
}
`

	if err := os.WriteFile(appDelegatePath, []byte(swiftContent), 0644); err != nil {
		return "", fmt.Errorf("failed to write AppDelegate.swift: %w", err)
	}

	return appDelegatePath, nil
}

func (p *IOSPlatform) createObjCAppDelegate(projectPath string) (string, error) {
	appDelegateHPath := filepath.Join(projectPath, "AppDelegate.h")
	appDelegateMPath := filepath.Join(projectPath, "AppDelegate.m")

	// Check if files already exist
	if fileExists(appDelegateMPath) {
		return appDelegateMPath, nil
	}

	// Create directory if needed
	if err := os.MkdirAll(projectPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory %s: %w", projectPath, err)
	}

	// Create AppDelegate.h
	headerContent := `#import <UIKit/UIKit.h>

@interface AppDelegate : UIResponder <UIApplicationDelegate>

@property (strong, nonatomic) UIWindow *window;

@end
`

	if err := os.WriteFile(appDelegateHPath, []byte(headerContent), 0644); err != nil {
		return "", fmt.Errorf("failed to write AppDelegate.h: %w", err)
	}

	// Create AppDelegate.m
	implementationContent := `#import "AppDelegate.h"
@import FirebaseCore;

@implementation AppDelegate

- (BOOL)application:(UIApplication *)application didFinishLaunchingWithOptions:(NSDictionary *)launchOptions {
    // Configure Firebase
    [FIRApp configure];
    
    // Override point for customization after application launch.
    return YES;
}

#pragma mark - UISceneSession lifecycle (iOS 13+)

- (UISceneConfiguration *)application:(UIApplication *)application 
                 configurationForConnectingSceneSession:(UISceneSession *)connectingSceneSession 
                                                options:(UISceneConnectionOptions *)options API_AVAILABLE(ios(13.0)) {
    return [[UISceneConfiguration alloc] initWithName:@"Default Configuration" sessionRole:connectingSceneSession.role];
}

- (void)application:(UIApplication *)application 
    didDiscardSceneSessions:(NSSet<UISceneSession *> *)sceneSessions API_AVAILABLE(ios(13.0)) {
    // Called when the user discards a scene session.
}

@end
`

	if err := os.WriteFile(appDelegateMPath, []byte(implementationContent), 0644); err != nil {
		return "", fmt.Errorf("failed to write AppDelegate.m: %w", err)
	}

	return appDelegateMPath, nil
}

func (p *IOSPlatform) isSwiftUIProject() bool {
	// Check for SwiftUI App files
	appFile := findFile(".", "*App.swift")
	if appFile != "" {
		content, err := os.ReadFile(appFile)
		if err == nil {
			contentStr := string(content)
			if strings.Contains(contentStr, "import SwiftUI") && strings.Contains(contentStr, "@main") {
				return true
			}
		}
	}

	// Check for SwiftUI imports in any Swift files
	swiftFiles := []string{}
	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if strings.HasSuffix(path, ".swift") {
			swiftFiles = append(swiftFiles, path)
		}
		return nil
	})

	if err == nil {
		for _, file := range swiftFiles {
			content, err := os.ReadFile(file)
			if err == nil {
				contentStr := string(content)
				if strings.Contains(contentStr, "import SwiftUI") {
					return true
				}
			}
		}
	}

	return false
}

func (p *IOSPlatform) createSwiftUIAppDelegateIntegration(projectPath string) (string, error) {
	// First create the AppDelegate
	appDelegatePath := filepath.Join(projectPath, "AppDelegate.swift")

	// Check if AppDelegate already exists
	if fileExists(appDelegatePath) {
		return appDelegatePath, nil
	}

	// Create directory if needed
	if err := os.MkdirAll(projectPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory %s: %w", projectPath, err)
	}

	// Create AppDelegate for SwiftUI
	swiftUIAppDelegateContent := `import UIKit
import FirebaseCore

class AppDelegate: NSObject, UIApplicationDelegate {
    
    func application(_ application: UIApplication,
                     didFinishLaunchingWithOptions launchOptions: [UIApplication.LaunchOptionsKey: Any]?) -> Bool {
        // Configure Firebase
        FirebaseApp.configure()
        
        // Override point for customization after application launch.
        return true
    }
    
    // MARK: UISceneSession Lifecycle (iOS 13+)
    func application(_ application: UIApplication,
                     configurationForConnecting connectingSceneSession: UISceneSession,
                     options: UIScene.ConnectionOptions) -> UISceneConfiguration {
        return UISceneConfiguration(name: "Default Configuration", sessionRole: connectingSceneSession.role)
    }
    
    func application(_ application: UIApplication, didDiscardSceneSessions sceneSessions: Set<UISceneSession>) {
        // Called when the user discards a scene session.
    }
}
`

	if err := os.WriteFile(appDelegatePath, []byte(swiftUIAppDelegateContent), 0644); err != nil {
		return "", fmt.Errorf("failed to write AppDelegate.swift: %w", err)
	}

	// Now find or create the SwiftUI App file and add the delegate adaptor
	if err := p.addDelegateAdaptorToSwiftUIApp(projectPath); err != nil {
		ui.WarningMsg(fmt.Sprintf("Created AppDelegate but failed to add delegate adaptor: %v", err))
		ui.InfoMsg("Please manually add '@UIApplicationDelegateAdaptor(AppDelegate.self) var delegate' to your SwiftUI App")
	}

	return appDelegatePath, nil
}

func (p *IOSPlatform) addDelegateAdaptorToSwiftUIApp(projectPath string) error {
	// Find existing SwiftUI App file
	appFile := findFile(".", "*App.swift")

	if appFile == "" {
		// Create a new SwiftUI App file
		return p.createSwiftUIAppFile(projectPath)
	}

	// Add delegate adaptor to existing App file
	content, err := os.ReadFile(appFile)
	if err != nil {
		return fmt.Errorf("failed to read App file: %w", err)
	}

	contentStr := string(content)

	// Check if delegate adaptor already exists
	if strings.Contains(contentStr, "@UIApplicationDelegateAdaptor") {
		ui.InfoMsg("UIApplicationDelegateAdaptor already exists in SwiftUI App")
		return nil
	}

	// Add delegate adaptor after @main line
	if strings.Contains(contentStr, "@main") {
		// Find the struct declaration
		lines := strings.Split(contentStr, "\n")
		var newLines []string
		delegateAdded := false

		for i := 0; i < len(lines); i++ {
			line := lines[i]
			newLines = append(newLines, line)

			// Add delegate adaptor after struct declaration
			if !delegateAdded && strings.Contains(line, "struct") && strings.Contains(line, "App") {
				// Look for the opening brace
				if strings.Contains(line, "{") {
					newLines = append(newLines, "    @UIApplicationDelegateAdaptor(AppDelegate.self) var delegate")
					newLines = append(newLines, "")
					delegateAdded = true
				} else if i+1 < len(lines) && strings.Contains(lines[i+1], "{") {
					// Opening brace is on next line
					i++ // Skip the next line since we're processing it here
					newLines = append(newLines, lines[i])
					newLines = append(newLines, "    @UIApplicationDelegateAdaptor(AppDelegate.self) var delegate")
					newLines = append(newLines, "")
					delegateAdded = true
				}
			}
		}

		if delegateAdded {
			newContent := strings.Join(newLines, "\n")
			if err := os.WriteFile(appFile, []byte(newContent), 0644); err != nil {
				return fmt.Errorf("failed to update SwiftUI App file: %w", err)
			}
			ui.SuccessMsg(fmt.Sprintf("Added UIApplicationDelegateAdaptor to: %s", appFile))
		} else {
			return fmt.Errorf("could not find appropriate location to add delegate adaptor")
		}
	}

	return nil
}

func (p *IOSPlatform) createSwiftUIAppFile(projectPath string) error {
	projectName := p.findProjectName()
	if projectName == "" {
		projectName = "MyApp"
	}

	appFileName := fmt.Sprintf("%sApp.swift", projectName)
	appFilePath := filepath.Join(projectPath, appFileName)

	// Check if file already exists
	if fileExists(appFilePath) {
		return nil
	}

	swiftUIAppContent := fmt.Sprintf(`import SwiftUI

@main
struct %sApp: App {
    @UIApplicationDelegateAdaptor(AppDelegate.self) var delegate
    
    var body: some Scene {
        WindowGroup {
            ContentView()
        }
    }
}
`, projectName)

	if err := os.WriteFile(appFilePath, []byte(swiftUIAppContent), 0644); err != nil {
		return fmt.Errorf("failed to write SwiftUI App file: %w", err)
	}

	// Also create a basic ContentView if it doesn't exist
	contentViewPath := filepath.Join(projectPath, "ContentView.swift")
	if !fileExists(contentViewPath) {
		contentViewContent := `import SwiftUI

struct ContentView: View {
    var body: some View {
        VStack {
            Image(systemName: "globe")
                .imageScale(.large)
                .foregroundColor(.accentColor)
            Text("Hello, world!")
        }
        .padding()
    }
}

struct ContentView_Previews: PreviewProvider {
    static var previews: some View {
        ContentView()
    }
}
`
		_ = os.WriteFile(contentViewPath, []byte(contentViewContent), 0644)
	}

	ui.SuccessMsg(fmt.Sprintf("Created SwiftUI App file: %s", appFilePath))
	return nil
}

func (p *IOSPlatform) runPackageManagerCommands() error {
	ui.InfoMsg("Running package manager commands...")

	// Check for CocoaPods first (preferred if Podfile exists)
	if p.findPodfile() != "" {
		return p.runPodInstall()
	}

	// Check for Swift Package Manager
	if p.hasSwiftPackages() {
		return p.updateSwiftPackages()
	}

	// If no package manager is detected, provide guidance
	ui.InfoMsg("📦 No package manager detected")
	ui.InfoMsg("")
	ui.InfoMsg("To proceed with Firebase setup, please choose a dependency manager:")
	ui.InfoMsg("")
	ui.InfoMsg("Option 1 - CocoaPods:")
	ui.InfoMsg("  1. Install CocoaPods: sudo gem install cocoapods")
	ui.InfoMsg("  2. Create Podfile: pod init")
	ui.InfoMsg("  3. Run 'nativefire configure' again")
	ui.InfoMsg("")
	ui.InfoMsg("Option 2 - Swift Package Manager (recommended):")
	ui.InfoMsg("  1. Open your Xcode project")
	ui.InfoMsg("  2. Go to File → Add Package Dependencies...")
	ui.InfoMsg("  3. Add: https://github.com/firebase/firebase-ios-sdk")
	ui.InfoMsg("  4. Run 'nativefire configure' again")
	ui.InfoMsg("")
	return nil
}

func (p *IOSPlatform) runPodInstall() error {
	// Check if pod command exists with spinner
	checkSpinner := ui.NewDotsSpinner("Checking CocoaPods installation...")
	checkSpinner.Start()

	err := p.checkPodCommand()
	checkSpinner.Stop()

	if err != nil {
		ui.AnimatedError("CocoaPods not found")
		ui.WarningMsg("Please install CocoaPods and run 'pod install' manually")
		ui.InfoMsg("Install CocoaPods: sudo gem install cocoapods")
		return nil
	}

	ui.AnimatedSuccess("CocoaPods found")

	// Run pod install with spinner
	return ui.ShowLoader("Installing CocoaPods dependencies", func() error {
		if err := p.runCommand("pod", []string{"install"}, "Installing CocoaPods dependencies"); err != nil {
			ui.WarningMsg("Failed to run 'pod install'. Please run it manually")
			ui.InfoMsg("Run: pod install")
			return err
		}

		ui.SuccessMsg("CocoaPods dependencies installed successfully!")
		ui.InfoMsg("Make sure to open the .xcworkspace file in Xcode, not the .xcodeproj file")
		return nil
	})
}

func (p *IOSPlatform) hasSwiftPackages() bool {
	// Check for Package.swift or .swiftpm directory
	if fileExists("Package.swift") || fileExists(".swiftpm") {
		return true
	}

	// Check for Xcode project with SPM dependencies
	xcodeproj := findFile(".", "*.xcodeproj")
	if xcodeproj != "" {
		// Check for Package.resolved in project directory
		packageResolved := filepath.Join(xcodeproj, "project.xcworkspace/xcshareddata/swiftpm/Package.resolved")
		if fileExists(packageResolved) {
			return true
		}
	}

	return false
}

func (p *IOSPlatform) shouldUseSPM() bool {
	// Prefer SPM if we find an Xcode project but no Podfile
	xcodeproj := findFile(".", "*.xcodeproj")
	podfile := p.findPodfile()

	// Use SPM if we have an Xcode project but no Podfile
	return xcodeproj != "" && podfile == ""
}

func (p *IOSPlatform) updateSwiftPackages() error {
	ui.InfoMsg("📦 Swift Package Manager detected")
	ui.InfoMsg("")
	ui.InfoMsg("Please ensure Firebase iOS SDK is properly added to your project:")
	ui.InfoMsg("")
	
	if fileExists("Package.swift") {
		ui.InfoMsg("For Package.swift projects:")
		ui.InfoMsg("  1. Add Firebase dependency to Package.swift")
		ui.InfoMsg("  2. Run: swift package resolve")
		ui.InfoMsg("  3. Add FirebaseCore to your target dependencies")
	} else {
		ui.InfoMsg("For Xcode projects:")
		ui.InfoMsg("  1. Open your project in Xcode")
		ui.InfoMsg("  2. File → Add Package Dependencies...")
		ui.InfoMsg("  3. Add: https://github.com/firebase/firebase-ios-sdk")
		ui.InfoMsg("  4. Select FirebaseCore product")
		ui.InfoMsg("  5. Build your project to resolve dependencies")
	}
	
	ui.InfoMsg("")
	ui.SuccessMsg("Swift Package Manager setup guidance provided")
	return nil
}

func (p *IOSPlatform) setupSPMFirebase() error {
	ui.AnimatedHeader("Firebase iOS SDK Setup Required")
	fmt.Println()
	
	ui.Typewriter("Please add Firebase iOS SDK to your Xcode project using Swift Package Manager:", 30*time.Millisecond)
	fmt.Println()
	
	ui.InfoMsg("📋 Steps to follow:")
	steps := []string{
		"Open your Xcode project",
		"Go to File → Add Package Dependencies...",
		"Enter this URL: https://github.com/firebase/firebase-ios-sdk",
		"Select version 10.24.0 or later",
		"Add 'FirebaseCore' to your app target",
		"Build your project to ensure dependencies are resolved",
	}
	
	for i, step := range steps {
		time.Sleep(200 * time.Millisecond)
		ui.Step(i+1, step)
	}
	fmt.Println()
	
	// Ask user to confirm before proceeding with interactive prompt
	response := ui.PromptWithSpinner("Have you completed the above steps?", []string{
		"Yes - Continue with Firebase initialization code setup",
		"No - Skip code setup for now",
	})
	
	if response != "1" && strings.ToLower(response) != "yes" && strings.ToLower(response) != "y" {
		ui.InfoMsg("⏸️  Code setup skipped. Run 'nativefire configure' again after adding Firebase SDK.")
		ui.InfoMsg("💡 Reminder: Don't forget to add your GoogleService-Info.plist to your Xcode project!")
		return fmt.Errorf("user chose to skip code setup")
	}
	
	ui.AnimatedSuccess("Proceeding with Firebase initialization code setup")
	return nil
}

func (p *IOSPlatform) setupSPMPackageSwift() error {
	ui.InfoMsg("🔥 Firebase iOS SDK Setup Required")
	ui.InfoMsg("")
	ui.InfoMsg("Detected Package.swift project. Please add Firebase iOS SDK dependency:")
	ui.InfoMsg("")
	ui.InfoMsg("📋 Add this to your Package.swift dependencies:")
	ui.InfoMsg(`  .package(url: "https://github.com/firebase/firebase-ios-sdk", from: "10.24.0")`)
	ui.InfoMsg("")
	ui.InfoMsg("📋 Add FirebaseCore to your target dependencies:")
	ui.InfoMsg(`  .product(name: "FirebaseCore", package: "firebase-ios-sdk")`)
	ui.InfoMsg("")
	ui.InfoMsg("📋 Then run:")
	ui.InfoMsg("  swift package resolve")
	ui.InfoMsg("")

	// Ask user to confirm before proceeding
	ui.InfoMsg("🤔 Have you completed the above steps?")
	ui.InfoMsg("   Type 'yes' to continue with Firebase initialization code setup")
	ui.InfoMsg("   Type 'no' or press Enter to skip code setup for now")
	ui.InfoMsg("")

	var response string
	fmt.Print("Continue with code setup? (yes/no): ")
	fmt.Scanln(&response)

	response = strings.ToLower(strings.TrimSpace(response))
	if response != "yes" && response != "y" {
		ui.InfoMsg("⏸️  Code setup skipped. Run 'nativefire configure' again after adding Firebase SDK.")
		ui.InfoMsg("💡 Reminder: Don't forget to add your GoogleService-Info.plist to your project!")
		return fmt.Errorf("user chose to skip code setup")
	}

	ui.SuccessMsg("✅ Proceeding with Firebase initialization code setup...")
	return nil
}

func (p *IOSPlatform) checkPodCommand() error {
	cmd := exec.Command("which", "pod")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("CocoaPods not found")
	}
	return nil
}

func (p *IOSPlatform) runCommand(command string, args []string, description string) error {
	ui.InfoMsg(fmt.Sprintf("Running: %s %s", command, strings.Join(args, " ")))
	_ = description // Parameter kept for interface consistency

	cmd := exec.Command(command, args...)
	output, err := cmd.CombinedOutput()

	if err != nil {
		ui.WarningMsg(fmt.Sprintf("Command failed: %s", string(output)))
		return err
	}

	if len(output) > 0 {
		ui.InfoMsg(string(output))
	}

	return nil
}
