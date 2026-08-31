#!/usr/bin/env python3
import hashlib
import json
import os
import shutil
import subprocess
import sys

OUT_DIR = "."
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


def main():
    path = sys.argv[1] if len(sys.argv) > 1 else "."

    run_go(path, "mod", "vendor")

    vendor_dir = os.path.join(path, "vendor")
    os.rename(os.path.join(
		vendor_dir, "modules.txt"),
		os.path.join(OUT_DIR, MODULES_TXT),
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
                "dest": f"vendor/{dst}",
                "sha256": sha256sum(zip_path),
            }
        )

    with open(os.path.join(OUT_DIR, GO_MOD_YML), "w") as f:
        write_yaml(f, files)


if __name__ == "__main__":
    try:
        main()
    except (subprocess.CalledProcessError, OSError, RuntimeError) as e:
        print(e, file=sys.stderr)
        sys.exit(1)
