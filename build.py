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
    ["linux", "mips"],
    ["linux", "mipsle"],
    ["linux", "mips64"],
    ["linux", "mips64le"],
    ["windows", "386"],
    ["windows", "amd64"]
              ]

IP_NETWORK_SAMPLE_DICT = {"name": "ip_network_sample",
                          "url": "https://github.com/17mon/china_ip_list/raw/master/china_ip_list.txt"}
DOMAIN_SAMPLE_DICT = {"name": "domain_sample",
                      "url": "https://github.com/gfwlist/gfwlist/raw/master/gfwlist.txt"}


def download_file():
    for d in [IP_NETWORK_SAMPLE_DICT, DOMAIN_SAMPLE_DICT]:
        try:
            subprocess.check_call("wget -O" + d["name"] + " " + d["url"], shell=True)
        except subprocess.CalledProcessError:
            print("Get " + d["url"] + " failed.")


def go_build_zip():
    subprocess.check_call("go get -v github.com/shawn1m/overture/main", shell=True)
    for o, a in GO_OS_ARCH_LIST:
        zip_name = "overture-" + o + "-" + a
        binary_name = zip_name + (".exe" if o == "windows" else "")
        version = subprocess.check_output("git describe --tags", shell=True).decode()
        try:
            subprocess.check_call("GOOS=" + o + " GOARCH=" + a + " CGO_ENABLED=0" + " go build -ldflags \"-s -w " +
                                  "-X main.version=" + version + "\" -o " + binary_name + " main/main.go", shell=True)
            subprocess.check_call("zip " + zip_name + ".zip " + binary_name + " " + IP_NETWORK_SAMPLE_DICT["name"] + " " +
                                  DOMAIN_SAMPLE_DICT["name"] + " hosts_sample config.json", shell=True)
        except subprocess.CalledProcessError:
            print(o + " " + a + " failed.")


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
