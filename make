#!/bin/python3

import sys, toml, os, time

global config


def format_duration(duration: float) -> str:
    minutes = int(duration // 60)
    seconds = int(duration % 60)
    milliseconds = int((duration - int(duration)) * 1000)

    return f"{minutes} m {seconds}.{milliseconds} s"


def fatal(message: str) -> None:
    print("FATAL: " + message)
    exit(1)


def info(message: str) -> None:
    print(f"\033[36m{message}\033[0m")


def done(duration: float) -> None:
    print(f"\033[32mDone ({format_duration(duration)})\033[0m")


def mkdir(name: str) -> None:
    os.system("mkdir -p " + name)


def clean() -> None:
    os.system("rm -rf ./bin ./.make_package")


def build_binaries() -> None:
    info("Building binaries...")
    start = time.time()

    for target in config["build"]["platforms"]:
        goos = target.split("/")[0]
        goarch = target.split("/")[1]

        exe = ".exe" if goos == "windows" else ""

        os.system(f"GOOS={goos} GOARCH={goarch} go build {config['build']['flags']} -o bin/{config['name']}_{goos}_{goarch}{exe} {config['build']['target']}")
    
    done(time.time() - start)


def package() -> None:
    info("Packaging apt...")
    start = time.time()

    mkdir(".make_package")

    package_name = config["name"] + "_" + config["version"] + "-amd64"
    package_path = ".make_package/" + package_name

    mkdir(package_path + "/usr/bin")
    mkdir(package_path + "/DEBIAN")

    with open(package_path + "/DEBIAN/control", "x") as file:
        file.write(f"Package: {config['name']}\n" +
            f"Version: {config['version']}\n" +
            "Architecture: amd64\n" +
            f"Maintainer: {config['maintainer']['name']} <{config['maintainer']['email']}>\n" +
            f"Description: {config['description']}\n"
        )

    os.system(f"cp bin/{config['name']}_linux_amd64 {package_path}/usr/bin")
    os.system(f"dpkg-deb --build --root-owner-group {package_path}")
    os.system(f"mv {package_path}.deb bin")

    done(time.time() - start)


def target_package() -> None:
    clean()
    build_binaries()
    package()


def target_binaries() -> None:
    clean()
    build_binaries()


def target_all() -> None:
    clean()
    build_binaries()
    package()


if __name__ == "__main__":
    # Load config
    try:
        with open("make.toml", "r") as f:
            config = toml.load(f)
    except Exception as e:
        fatal("Failed to open make config: " + str(e))

    # Run action
    action = ""

    if len(sys.argv) > 1:
        action = sys.argv[1]

    if action == "":
        target_all()

    elif action == "clean":
        clean()
    elif action == "binary":
        target_binaries()
    elif action == "package":
        target_package()
