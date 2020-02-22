#!/usr/bin/env python3

import subprocess
import sys

GO_OS_ARCH_LIST = [
    ["darwin", "amd64"],
    ["linux", "386"],
    ["linux", "amd64"],
    ["linux", "arm"],
    ["linux", "arm64"],
    ["linux", "mips", "softfloat"],
    ["linux", "mips", "hardfloat"],
    ["linux", "mipsle", "softfloat"],
    ["linux", "mipsle", "hardfloat"],
    ["linux", "mips64"],
    ["linux", "mips64le"],
    ["freebsd", "386"],
    ["freebsd", "amd64"],
    ["windows", "386"],
    ["windows", "amd64"]
              ]

GO_IOS_ARCH_LIST = [
    ["darwin", "arm64"],
    ["darwin", "arm"]
]

GO_ANDROID_ARCH_LIST = [
    ["android", "arm", "arm-linux-androideabi"],
    ["android", "arm64", "aarch64-linux-android"],
    ["android", "386", "i686-linux-android"],
    ["android", "amd64", "x86_64-linux-android"],
]


def go_build_desktop(binary_name, version, o, a, p):
    mipsflag = (" GOMIPS=" + (p[0] if p else "") if p else "")
    subprocess.check_call("GOOS=" + o + " GOARCH=" + a + mipsflag + " CGO_ENABLED=0" + " go build -ldflags \"-s -w " +
                                  "-X main.version=" + version + "\" -o " + binary_name + " main/main.go", shell=True)

def go_build_ios(binary_name, version, o, a, p):
    subprocess.check_call("CC=$(go env GOROOT)/misc/ios/clangwrap.sh GOOS=" + o + " GOARCH=" + a + " CGO_ENABLED=1" + " go build -ldflags \"-s -w " +
                                  "-X main.version=" + version + "\" -o " + binary_name + " main/main.go", shell=True)

def go_build_android(binary_name, version, o, a, p):
    triple = p[0]
    subprocess.check_call("CC=$ANDROID_NDK_ROOT/bin/" + triple + "/bin/clang GOOS=" + o + " GOARCH=" + a + " CGO_ENABLED=1" + " go build -ldflags \"-s -w " +
                                  "-X main.version=" + version + "\" -o " + binary_name + " main/main.go", shell=True)

def go_build_zip(arches, builder):
    subprocess.check_call("GOOS=windows go get -v github.com/shawn1m/overture/main", shell=True)
    for o, a, *p in arches:
        zip_name = "overture-" + o + "-" + a + ("-" + (p[0] if p else "") if p else "")
        binary_name = zip_name + (".exe" if o == "windows" else "")
        version = subprocess.check_output("git describe --tags", shell=True).decode()
        try:
            builder(binary_name, version, o, a, p)
            subprocess.check_call("zip " + zip_name + ".zip " + binary_name + " " + "hosts_sample "
                                                                                    "ip_network_primary_sample "
                                                                                    "ip_network_alternative_sample "
                                                                                    "domain_primary_sample "
                                                                                    "domain_alternative_sample "
                                                                                    "domain_ttl_sample "
                                                                                    "config.json", shell=True)
        except subprocess.CalledProcessError:
            print(o + " " + a + " " + (p[0] if p else "") + " failed.")


def create_sample_file():
    with open("./hosts_sample", "w") as f:
        f.write("127.0.0.1 localhost")
    with open("./ip_network_primary_sample", "w") as f:
        f.write("127.0.0.9/32")
    with open("./ip_network_alternative_sample", "w") as f:
        f.write("127.0.0.10/32")
    with open("./domain_primary_sample", "w") as f:
        f.write("primary.example")
    with open("./domain_alternative_sample", "w") as f:
        f.write("alternative.example")
    with open("./domain_ttl_sample", "w") as f:
        f.write("ttl.example 1000")


if __name__ == "__main__":

    subprocess.check_call("cp config.sample.json config.json", shell=True)

    if "-create-sample" in sys.argv:
        create_sample_file()

    if "-build" in sys.argv:
        go_build_zip(GO_OS_ARCH_LIST, go_build_desktop)
    
    if "-build-ios" in sys.argv:
        go_build_zip(GO_IOS_ARCH_LIST, go_build_ios)
    
    if "-build-android" in sys.argv:
        go_build_zip(GO_ANDROID_ARCH_LIST, go_build_android)
