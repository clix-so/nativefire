class Nativefire < Formula
  desc "Simplify Firebase setup in native development environments"
  homepage "https://github.com/clix-so/nativefire"
  url "https://github.com/clix-so/nativefire/archive/v1.0.0.tar.gz"
  sha256 "REPLACE_WITH_ACTUAL_SHA256_HASH"
  license "MIT"

  depends_on "go" => :build

  def install
    system "go", "build", *std_go_args(ldflags: "-s -w -X github.com/clix-so/nativefire/cmd.Version=#{version}")
  end

  test do
    assert_match "nativefire v#{version}", shell_output("#{bin}/nativefire version")
  end
end