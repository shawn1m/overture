#!/usr/bin/env python3

import subprocess

go_os_arch_list = [
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

china_ip_list_dict = {"name": "china_ip_list.txt",
                      "url": "https://github.com/17mon/china_ip_list/raw/master/china_ip_list.txt"}
gfwlist_dict = {"name": "gfwlist.txt",
                "url": "https://github.com/gfwlist/gfwlist/raw/master/gfwlist.txt"}

if __name__ == "__main__":

    subprocess.check_call("cp config.sample.json config.json", shell=True)

    for url in [china_ip_list_dict["url"], gfwlist_dict["url"]]:
        try:
            subprocess.check_call("wget -N " + url, shell=True)
        except subprocess.CalledProcessError:
            print("Get " + url + " failed.")

    for o, a in go_os_arch_list:
        name = "overture-" + o + "-" + a
        version = subprocess.check_output("git describe --tags", shell=True).decode()
        try:
            subprocess.check_call("GOOS=" + o + " GOARCH=" + a + " CGO_ENABLED=0" +
                                  " go build -ldflags " + "\"-X main.version=" + version + "\" -o " + name + " main/main.go", shell=True)
            subprocess.check_call("zip " + name + ".zip " +
                                  name + " " + china_ip_list_dict["name"] + " " +
                                  gfwlist_dict["name"] + " hosts config.json", shell=True)
        except subprocess.CalledProcessError:
            print(o + " " + a + " failed.")