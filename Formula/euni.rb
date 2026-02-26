class Euni < Formula
  desc "Detect and remediate Unicode normalization drift in Git repos"
  homepage "https://github.com/takuto-tanaka-4digit/excel-unidiff-cli"
  url "https://github.com/takuto-tanaka-4digit/excel-unidiff-cli/archive/334e1a15eaed883e21e0ad58192fbe3c84443a2a.tar.gz"
  version "0.0.0-334e1a1"
  sha256 "f493ce6db032571abc356b4ce430309f65d174000757be00e0ab15da2463fb67"
  head "https://github.com/takuto-tanaka-4digit/excel-unidiff-cli.git", branch: "main"

  depends_on "go" => :build

  def install
    ldflags = %W[
      -s
      -w
      -X
      main.version=#{version}
      -X
      main.commit=334e1a1
    ]
    system "go", "build", *std_go_args(ldflags: ldflags), "./cmd/euni"
  end

  test do
    output = shell_output("#{bin}/euni version")
    assert_match "euni", output
  end
end
