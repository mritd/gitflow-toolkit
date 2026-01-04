class GitflowToolkit < Formula
  desc "CLI tool for standardizing git commits following Angular commit specification"
  homepage "https://github.com/mritd/gitflow-toolkit"
  license "MIT"

  # Get latest version from GitHub API
  latest_version = JSON.parse(
    URI.open("https://api.github.com/repos/mritd/gitflow-toolkit/releases/latest").read
  )["tag_name"].delete_prefix("v")

  version latest_version

  # Fetch checksums from release
  checksums_url = "https://github.com/mritd/gitflow-toolkit/releases/download/v#{version}/checksums.txt"
  checksums = URI.open(checksums_url).read.lines.to_h { |l| l.strip.split("  ").reverse }

  on_macos do
    on_intel do
      url "https://github.com/mritd/gitflow-toolkit/releases/download/v#{version}/gitflow-toolkit-darwin-amd64"
      sha256 checksums["gitflow-toolkit-darwin-amd64"]
    end

    on_arm do
      url "https://github.com/mritd/gitflow-toolkit/releases/download/v#{version}/gitflow-toolkit-darwin-arm64"
      sha256 checksums["gitflow-toolkit-darwin-arm64"]
    end
  end

  on_linux do
    on_intel do
      if Hardware::CPU.is_64_bit?
        url "https://github.com/mritd/gitflow-toolkit/releases/download/v#{version}/gitflow-toolkit-linux-amd64"
        sha256 checksums["gitflow-toolkit-linux-amd64"]
      else
        url "https://github.com/mritd/gitflow-toolkit/releases/download/v#{version}/gitflow-toolkit-linux-386"
        sha256 checksums["gitflow-toolkit-linux-386"]
      end
    end

    on_arm do
      if Hardware::CPU.is_64_bit?
        url "https://github.com/mritd/gitflow-toolkit/releases/download/v#{version}/gitflow-toolkit-linux-arm64"
        sha256 checksums["gitflow-toolkit-linux-arm64"]
      else
        url "https://github.com/mritd/gitflow-toolkit/releases/download/v#{version}/gitflow-toolkit-linux-armv7"
        sha256 checksums["gitflow-toolkit-linux-armv7"]
      end
    end
  end

  def install
    binary_name = stable.url.split("/").last
    bin.install binary_name => "gitflow-toolkit"

    # Create git subcommand symlinks
    %w[ci ps feat fix docs style refactor test chore perf hotfix].each do |cmd|
      bin.install_symlink "gitflow-toolkit" => "git-#{cmd}"
    end
  end

  def caveats
    <<~EOS
      Git subcommands have been installed:
        git ci       - Interactive commit
        git ps       - Push with TUI
        git feat     - Create feature branch
        git fix      - Create fix branch
        ... and more

      For AI-powered commit generation, configure your LLM:
        git config --global gitflow.llm-api-key "your-api-key"
    EOS
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/gitflow-toolkit -v")
  end
end
