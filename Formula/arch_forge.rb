class ArchForge < Formula
  desc "Generate Go project structures based on proven architectural patterns"
  homepage "https://github.com/archforge/cli"
  license "MIT"

  on_macos do
    on_arm do
      url "https://github.com/archforge/cli/releases/download/v#{version}/arch_forge_#{version}_darwin_arm64.tar.gz"
      sha256 "GORELEASERPLACEHOLDER"
    end
    on_intel do
      url "https://github.com/archforge/cli/releases/download/v#{version}/arch_forge_#{version}_darwin_amd64.tar.gz"
      sha256 "GORELEASERPLACEHOLDER"
    end
  end

  on_linux do
    on_arm do
      url "https://github.com/archforge/cli/releases/download/v#{version}/arch_forge_#{version}_linux_arm64.tar.gz"
      sha256 "GORELEASERPLACEHOLDER"
    end
    on_intel do
      url "https://github.com/archforge/cli/releases/download/v#{version}/arch_forge_#{version}_linux_amd64.tar.gz"
      sha256 "GORELEASERPLACEHOLDER"
    end
  end

  def install
    bin.install "arch_forge"
  end

  test do
    system "#{bin}/arch_forge", "--version"
  end
end
