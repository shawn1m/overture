#!/usr/bin/env python3

import subprocess
import base64
import sys

GO_OS_ARCH_LIST = [
    ["darwin", "amd64"],
    ["linux", "386"],
    ["linux", "amd64"],
    ["linux", "arm"],
    ["linux", "arm64"],
    ["linux", "mips", "softfloat"],
    ["linux", "mips", "hardfloat"],
    ["linux", "mipsle"],
    ["linux", "mips64"],
    ["linux", "mips64le"],
    ["windows", "386"],
    ["windows", "amd64"]
              ]

IP_NETWORK_SAMPLE_DICT = {"name": "ip_network_sample",
                          "url": "https://github.com/17mon/china_ip_list/raw/master/china_ip_list.txt",
                          "parse": ""}
DOMAIN_SAMPLE_DICT = {"name": "domain_sample",
                      "url": "https://github.com/gfwlist/gfwlist/raw/master/gfwlist.txt",
                      "parse": ""}

DOMAIN_WHITE_LIST_DICT = {"name": "apple.china.conf",
                          "url": "https://raw.githubusercontent.com/felixonmars/dnsmasq-china-list/master/apple.china.conf",
                          "parse": "cat apple.china.conf | awk '{FS=\"/\"}{print $2}' > domain_white_sample "}


def download_file():
    for d in [IP_NETWORK_SAMPLE_DICT, DOMAIN_SAMPLE_DICT, DOMAIN_WHITE_LIST_DICT]:
        try:
            subprocess.check_call("wget -O" + d["name"] + " " + d["url"], shell=True)
        except subprocess.CalledProcessError:
            print("Get " + d["url"] + " failed.")

        if d["parse"] != "":
            subprocess.check_call(d["parse"], shell=True)


def go_build_zip():
    subprocess.check_call("go get -v github.com/shawn1m/overture/main", shell=True)
    for o, a, *p in GO_OS_ARCH_LIST:
        zip_name = "overture-" + o + "-" + a + ("-" + (p[0] if p else "") if p else "")
        binary_name = zip_name + (".exe" if o == "windows" else "")
        version = subprocess.check_output("git describe --tags", shell=True).decode()
        mipsflag = (" GOMIPS=" + (p[0] if p else "") if p else "")
        try:
            subprocess.check_call("GOOS=" + o + " GOARCH=" + a + mipsflag + " CGO_ENABLED=0" + " go build -ldflags \"-s -w " +
                                  "-X main.version=" + version + "\" -o " + binary_name + " main/main.go", shell=True)
            subprocess.check_call("zip " + zip_name + ".zip " + binary_name + " " + IP_NETWORK_SAMPLE_DICT["name"] + " " +
                                  DOMAIN_SAMPLE_DICT["name"] + " hosts_sample domain_white_sample config.json", shell=True)
        except subprocess.CalledProcessError:
            print(o + " " + a + " " + (p[0] if p else "") + " failed.")


def decode_domain_sample():
    with open("./domain_sample", "rb") as fr:
        with open("./domain_temp", "w") as fw:
            file_decoded = base64.b64decode(fr.read()).decode()
            i = file_decoded.index("Whitelist Start")
            fw.write(file_decoded[:i])
    subprocess.check_call("mv domain_temp domain_sample", shell=True)


def create_hosts_sample_file():
    with open("./hosts_sample", "w") as f:
        f.write("127.0.0.1 localhost")


if __name__ == "__main__":

    subprocess.check_call("cp config.sample.json config.json", shell=True)

    if "-id" not in sys.argv:
        download_file()
        # decode_domain_sample()
        create_hosts_sample_file()

    if "-ib" not in sys.argv:
        go_build_zip()
