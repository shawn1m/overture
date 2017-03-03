#!/usr/bin/env python3

import subprocess

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

CHINA_IP_LIST_DICT = {"name": "china_ip_list.txt",
                      "url": "https://github.com/17mon/china_ip_list/raw/master/china_ip_list.txt"}
GFWLIST_DICT = {"name": "gfwlist.txt",
                "url": "https://github.com/gfwlist/gfwlist/raw/master/gfwlist.txt"}

if __name__ == "__main__":

    subprocess.check_call("cp config.sample.json config.json", shell=True)

    for url in [CHINA_IP_LIST_DICT["url"], GFWLIST_DICT["url"]]:
        try:
            subprocess.check_call("wget -N " + url, shell=True)
        except subprocess.CalledProcessError:
            print("Get " + url + " failed.")

    for o, a in GO_OS_ARCH_LIST:
        zip_name = "overture-" + o + "-" + a
        binary_name = zip_name + (".exe" if o == "windows" else "")
        version = subprocess.check_output("git describe --tags", shell=True).decode()
        try:
            subprocess.check_call("GOOS=" + o + " GOARCH=" + a + " CGO_ENABLED=0" +
                                  " go build -ldflags " + "\"-X main.version=" + version + "\" -o " + binary_name + " main/main.go", shell=True)
            subprocess.check_call("zip " + zip_name + ".zip " +
                                  binary_name + " " + CHINA_IP_LIST_DICT["name"] + " " +
                                  GFWLIST_DICT["name"] + " hosts config.json", shell=True)
        except subprocess.CalledProcessError:
            print(o + " " + a + " failed.")