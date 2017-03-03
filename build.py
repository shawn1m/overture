#!/usr/bin/env python3

import subprocess

GO_OS_ARCH_LIST = [
        ["darwin", "amd64", "0"],
        ["linux", "386", "0"],
        ["linux", "amd64", "0"],
        ["linux", "mips", "0"],
        ["linux", "mipsle", "0"],
        ["linux", "mips64", "0"],
        ["linux", "mips64le", "0"],
        ["windows", "386", "0"],
        ["windows", "amd64", "0"],
        ["linux", "arm64", "0"],
        ["linux", "arm", "5"],
        ["linux", "arm", "6"],
        ["linux", "arm", "7"]
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

    for GOOS, GOARCH, GOARM in GO_OS_ARCH_LIST:
        if GOARM == '0':
            if GOOS == 'windows':
                name = "overture-" + GOOS + "-" + GOARCH + '.exe'
                version = subprocess.check_output("git describe --tags", shell=True).decode()
                try:
                    subprocess.check_call("GOOS=" + GOOS + " GOARCH=" + GOARCH + " CGO_ENABLED=0" +
                                          " go build -ldflags " + "\"-X main.version=" + version + "\" -o " + name + " main/main.go", shell=True)
                    subprocess.check_call("zip " + name + ".zip " +
                                          name + " " + china_ip_list_dict["name"] + " " +
                                          gfwlist_dict["name"] + " hosts config.json", shell=True)
                except subprocess.CalledProcessError:
                    print(GOOS + " " + GOARCH + " failed.")
                noguiName = "noGUI-overture-" + GOOS + "-" + GOARCH + '.exe'
                version = subprocess.check_output("git describe --tags", shell=True).decode()
                try:
                    subprocess.check_call("GOOS=" + GOOS + " GOARCH=" + GOARCH + " CGO_ENABLED=0" +
                                          " go build -ldflags " + "\'-H windowsgui -X main.version=" + version + "\' -o " + noguiName + " main/main.go", shell=True)
                    subprocess.check_call("zip " + noguiName + ".zip " +
                                          noguiName + " " + china_ip_list_dict["name"] + " " +
                                          gfwlist_dict["name"] + " hosts config.json", shell=True)
                except subprocess.CalledProcessError:
                    print(GOOS + " " + GOARCH + " failed.")
            else:
                name = "overture-" + GOOS + "-" + GOARCH
                version = subprocess.check_output("git describe --tags", shell=True).decode()
                try:
                    subprocess.check_call("GOOS=" + GOOS + " GOARCH=" + GOARCH + " CGO_ENABLED=0" +
                                          " go build -ldflags " + "\"-X main.version=" + version + "\" -o " + name + " main/main.go", shell=True)
                    subprocess.check_call("zip " + name + ".zip " +
                                          name + " " + china_ip_list_dict["name"] + " " +
                                          gfwlist_dict["name"] + " hosts config.json", shell=True)
                except subprocess.CalledProcessError:
                    print(GOOS + " " + GOARCH + " failed.")
        else:
            name = "overture-" + GOOS + "-" + GOARCH + "-" + GOARM
            version = subprocess.check_output("git describe --tags", shell=True).decode()
            try:
                subprocess.check_call("GOOS=" + GOOS + " GOARCH=" + GOARCH + " GOARM=" + GOARM +" CGO_ENABLED=0" +
                                      " go build -ldflags " + "\"-X main.version=" + version + "\" -o " + name + " main/main.go", shell=True)
                subprocess.check_call("zip " + name + ".zip " +
                                      name + " " + china_ip_list_dict["name"] + " " +
                                      gfwlist_dict["name"] + " hosts config.json", shell=True)
            except subprocess.CalledProcessError:
                print(GOOS + " " + GOARCH + " " + GOARM + " failed.")
