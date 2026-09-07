#!/usr/bin/env python3
import argparse
import hashlib
import json
import os
import shutil
import subprocess
import sys

MODULES_TXT = "modules.txt"
GO_MOD_YML = "go.mod.yml"

def run_go(path, *args, capture=False):
    proc = subprocess.run(
        ["go", *args],
        cwd=path,
        check=True,
        stdout=subprocess.PIPE if capture else None,
        text=True,
    )
    return proc.stdout


def go_mod_download_json(path):
    out = run_go(path, "mod", "download", "-json", capture=True)
    dec = json.JSONDecoder()
    mods = []
    i = 0
    n = len(out)
    while True:
        while i < n and out[i].isspace():
            i += 1
        if i >= n:
            break
        m, i = dec.raw_decode(out, i)
        mods.append(m)
    return mods


def go_mod_replacements(path):
    data = json.loads(run_go(path, "mod", "edit", "-json", capture=True))
    return {r["New"]["Path"]: r["Old"]["Path"]
        for r in data.get("Replace") or []}


def sha256sum(path):
    h = hashlib.sha256()
    with open(path, "rb") as f:
        for chunk in iter(lambda: f.read(1 << 20), b""):
            h.update(chunk)
    return h.hexdigest()


def write_yaml(f, files):
    for entry in files:
        first = True
        for key, val in entry.items():
            prefix = "- " if first else "  "
            f.write(f"{prefix}{key}: {val}\n")
            first = False


def parse_args():
    parser = argparse.ArgumentParser(
        description="Generate flatpak-builder sources for Go dependencies")
    parser.add_argument("path", nargs="?", default=".",
        help="Go module directory (default: current directory)")
    parser.add_argument("--dest-prefix", default="",
        help="directory prepended to vendor dest paths, use the module "
            "subdir when the manifest checks out the whole repository")
    parser.add_argument("--out-dir", default=None,
        help="directory for generated files (default: module directory)")
    parser.add_argument("--name", default="",
        help="prefix for generated file names, e.g. 'service' writes "
            "service-go.mod.yml and service-modules.txt")
    return parser.parse_args()


def main():
    args = parse_args()
    path = args.path
    out_dir = args.out_dir if args.out_dir is not None else path
    name = args.name + "-" if args.name else ""
    dest_prefix = args.dest_prefix.strip("/")
    vendor_dest = "vendor" if not dest_prefix else dest_prefix + "/vendor"

    os.makedirs(out_dir, exist_ok=True)

    run_go(path, "mod", "vendor")

    vendor_dir = os.path.join(path, "vendor")
    shutil.move(
        os.path.join(vendor_dir, "modules.txt"),
        os.path.join(out_dir, name + MODULES_TXT),
    )
    shutil.rmtree(vendor_dir)

    mods = go_mod_download_json(path)
    replaced = go_mod_replacements(path)

    files = []

    for m in mods:
        mod_path = m["Path"]
        zip_path = m.get("Zip")
        if not zip_path:
            raise RuntimeError(
                f"no zip file for module {mod_path}@{m.get('Version')}")

        dst = replaced.get(mod_path, mod_path)
        i = zip_path.find("download")
        if i < 0:
            raise RuntimeError(
                f"unsupported zip file path: {zip_path!r}")

        files.append(
            {
                "type": "archive",
                "url": "https://proxy.golang.org/" +
                    zip_path[i + len("download/"):],
                "strip-components": mod_path.count("/") + 1,
                "dest": f"{vendor_dest}/{dst}",
                "sha256": sha256sum(zip_path),
            }
        )

    with open(os.path.join(out_dir, name + GO_MOD_YML), "w") as f:
        write_yaml(f, files)


if __name__ == "__main__":
    try:
        main()
    except (subprocess.CalledProcessError, OSError, RuntimeError) as e:
        print(e, file=sys.stderr)
        sys.exit(1)
