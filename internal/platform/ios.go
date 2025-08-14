package platform

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

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
		return fmt.Errorf("failed to write config file to %s: %w", targetPath, err)
	}

	// Clean up the temp file after successful installation
	if config.ConfigFile != "" {
		os.Remove(config.ConfigFile)
	}

	ui.SuccessMsg(fmt.Sprintf("Configuration file installed at: %s", targetPath))
	ui.WarningMsg("Please add GoogleService-Info.plist to your Xcode project manually")
	return nil
}

func (p *IOSPlatform) AddInitializationCode(config *firebase.Config) error {
	podfilePath := p.findPodfile()
	podsAdded := false

	if podfilePath != "" {
		if err := p.addFirebasePods(podfilePath); err != nil {
			return err
		}
		podsAdded = true
	}

	appDelegatePath := p.findAppDelegate()
	if appDelegatePath == "" {
		// AppDelegate not found, create one
		ui.InfoMsg("AppDelegate not found, creating one...")
		var err error
		appDelegatePath, err = p.createAppDelegate()
		if err != nil {
			ui.WarningMsg(fmt.Sprintf("Failed to create AppDelegate: %v", err))
			ui.WarningMsg("Please manually add Firebase initialization code")
			return nil
		}
		ui.SuccessMsg(fmt.Sprintf("Created AppDelegate at: %s", appDelegatePath))
	}

	if err := p.addFirebaseInitialization(appDelegatePath); err != nil {
		return err
	}

	// Run package manager commands after all changes
	if podsAdded {
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

	// Add Firebase import
	if !strings.Contains(contentStr, "import Firebase") {
		if strings.Contains(contentStr, "import UIKit") {
			contentStr = strings.Replace(contentStr,
				"import UIKit",
				"import UIKit\nimport Firebase", 1)
		} else {
			// Add import at the beginning
			contentStr = "import Firebase\n" + contentStr
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

	// Add Firebase import
	if !strings.Contains(contentStr, "@import Firebase;") &&
		!strings.Contains(contentStr, "#import <Firebase/Firebase.h>") {
		if strings.Contains(contentStr, "#import \"AppDelegate.h\"") {
			contentStr = strings.Replace(contentStr,
				"#import \"AppDelegate.h\"",
				"#import \"AppDelegate.h\"\n@import Firebase;", 1)
		} else {
			// Add import at the beginning
			contentStr = "@import Firebase;\n" + contentStr
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

@main
class AppDelegate: UIResponder, UIApplicationDelegate {

    var window: UIWindow?

    func application(_ application: UIApplication,
                     didFinishLaunchingWithOptions launchOptions: [UIApplication.LaunchOptionsKey: Any]?) -> Bool {
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

@implementation AppDelegate

- (BOOL)application:(UIApplication *)application didFinishLaunchingWithOptions:(NSDictionary *)launchOptions {
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

class AppDelegate: NSObject, UIApplicationDelegate {
    
    func application(_ application: UIApplication,
                     didFinishLaunchingWithOptions launchOptions: [UIApplication.LaunchOptionsKey: Any]?) -> Bool {
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

	// Check for CocoaPods
	if p.findPodfile() != "" {
		return p.runPodInstall()
	}

	// Check for Swift Package Manager
	if p.hasSwiftPackages() {
		return p.updateSwiftPackages()
	}

	ui.InfoMsg("No package manager detected")
	return nil
}

func (p *IOSPlatform) runPodInstall() error {
	ui.InfoMsg("Installing CocoaPods dependencies...")

	// Check if pod command exists
	if err := p.checkPodCommand(); err != nil {
		ui.WarningMsg("CocoaPods not found. Please install CocoaPods and run 'pod install' manually")
		ui.InfoMsg("Install CocoaPods: sudo gem install cocoapods")
		return nil
	}

	// Run pod install
	if err := p.runCommand("pod", []string{"install"}, "Installing CocoaPods dependencies"); err != nil {
		ui.WarningMsg("Failed to run 'pod install'. Please run it manually")
		ui.InfoMsg("Run: pod install")
		return nil
	}

	ui.SuccessMsg("CocoaPods dependencies installed successfully!")
	ui.InfoMsg("Make sure to open the .xcworkspace file in Xcode, not the .xcodeproj file")
	return nil
}

func (p *IOSPlatform) hasSwiftPackages() bool {
	// Check for Package.swift or .swiftpm directory
	return fileExists("Package.swift") || fileExists(".swiftpm")
}

func (p *IOSPlatform) updateSwiftPackages() error {
	ui.InfoMsg("Updating Swift Package dependencies...")

	// For SPM projects, we can try to resolve packages
	if err := p.runCommand("swift", []string{"package", "resolve"}, "Resolving Swift Package dependencies"); err != nil {
		ui.WarningMsg("Failed to resolve Swift packages. Please update them manually in Xcode")
		ui.InfoMsg("In Xcode: File > Package Dependencies > Reset Package Caches")
		return nil
	}

	ui.SuccessMsg("Swift Package dependencies resolved successfully!")
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
