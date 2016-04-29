require "language/go"

class Sup < Formula
  desc "Stack Up. Super simple deployment tool - think of it like 'make' for a network of servers."
  homepage "https://github.com/pressly/sup"
  url "https://github.com/pressly/sup/archive/4ee5083c8321340bc2a6410f24d8a760f7ad3847.zip"
  version "0.3.1"
  sha256 "7fa17c20fdcd9e24d8c2fe98081e1300e936da02b3f2cf9c5a11fd699cbc487e"

  depends_on "go"  => :build

  def install
    ENV["GOBIN"] = bin
    ENV["GOPATH"] = buildpath
    ENV["GOHOME"] = buildpath

    mkdir_p buildpath/"src/github.com/pressly/"
    ln_sf buildpath, buildpath/"src/github.com/pressly/sup"
    Language::Go.stage_deps resources, buildpath/"src"

    system "go", "build", "-o", bin/"sup", "./cmd/sup"
  end

  test do
    assert_equal "0.3", shell_output("#{bin}/bin/sup")
  end
end
