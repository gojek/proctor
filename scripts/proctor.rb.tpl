class Proctor < Formula
  desc "Proctor CLI"
  homepage "https://github.com/gojektech/proctor"
  url "https://github.com/gojek/proctor/releases/download/v{{ .Tag }}/proctor_{{ .Tag }}_Darwin_x86_64.tar.gz"
  version "{{ .Tag }}"
  sha256 "{{ .SHA }}"
  head "https://github.com/gojek/proctor.git"

  bottle :unneeded

  def install
    bin.install "proctor"
  end

  test do
    system "#{bin}/proctor", "--help"
  end
end
