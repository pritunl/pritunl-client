import datetime
import re
import sys
import subprocess
import time
import math
import json
import requests
import os
import zlib
import getpass
import base64
from cryptography.hazmat.primitives import hashes
from cryptography.hazmat.backends import default_backend
from cryptography.hazmat.primitives.ciphers import (
    Cipher, algorithms, modes
)
from cryptography.hazmat.primitives.kdf.pbkdf2 import PBKDF2HMAC

CHANGES_PATH = 'CHANGES'
CONSTANTS_PATH = 'service/constants/constants.go'
CONSTANTS_PATH2 = 'client/package.json'
CONSTANTS_PATH3 = 'client/package-lock.json'
CONSTANTS_PATH4 = 'cli/constants/constants.go'
CONSTANTS_PATH5 = 'resources_win/setup.iss'
STABLE_PACUR_PATH = '../pritunl-pacur'
TEST_PACUR_PATH = '../pritunl-pacur-test'
BUILD_KEYS_PATH = os.path.expanduser('~/data/build/pritunl_build.json')
BUILD_TARGETS = ('pritunl-client-electron', 'pritunl-client')
REPO_NAME = 'pritunl-client'
RELEASE_NAME = 'Pritunl Client'

cur_date = datetime.datetime.utcnow()
cmd = sys.argv[1]

def aes_encrypt(passphrase, data):
    enc_salt = os.urandom(32)
    enc_iv = os.urandom(12)

    kdf = PBKDF2HMAC(
        algorithm=hashes.SHA512(),
        length=32,
        salt=enc_salt,
        iterations=10000000,
        backend=default_backend(),
    )
    enc_key = kdf.derive(passphrase.encode())

    cipher = Cipher(
        algorithms.AES(enc_key),
        modes.GCM(enc_iv),
        backend=default_backend()
    ).encryptor()

    enc_data = cipher.update(data.encode('utf-8')) + cipher.finalize()
    auth_tag = cipher.tag

    return '\n'.join([
        base64.b64encode(enc_salt).decode('utf-8'),
        base64.b64encode(enc_iv).decode('utf-8'),
        base64.b64encode(enc_data).decode('utf-8'),
        base64.b64encode(auth_tag).decode('utf-8'),
    ])

def aes_decrypt(passphrase, data):
    data = data.split('\n')
    if len(data) < 4:
        raise ValueError('Invalid encryption data')

    enc_salt = base64.b64decode(data[0])
    enc_iv = base64.b64decode(data[1])
    enc_data = base64.b64decode(data[2])
    auth_tag = base64.b64decode(data[3])

    kdf = PBKDF2HMAC(
        algorithm=hashes.SHA512(),
        length=32,
        salt=enc_salt,
        iterations=10000000,
        backend=default_backend(),
    )
    enc_key = kdf.derive(passphrase.encode())

    cipher = Cipher(
        algorithms.AES(enc_key),
        modes.GCM(enc_iv, auth_tag),
        backend=default_backend()
    ).decryptor()

    decrypted_data = cipher.update(enc_data) + cipher.finalize()

    return decrypted_data.decode('utf-8')

passphrase = getpass.getpass('Enter passphrase: ')

if cmd == 'encrypt':
    passphrase2 = getpass.getpass('Enter passphrase: ')

    if passphrase != passphrase2:
        print('ERROR: Passphrase mismatch')
        sys.exit(1)

    with open(BUILD_KEYS_PATH, 'r') as build_keys_file:
        data = build_keys_file.read().strip()

    enc_data = aes_encrypt(passphrase, data)

    with open(BUILD_KEYS_PATH, 'w') as build_keys_file:
        build_keys_file.write(enc_data)

    sys.exit(0)

if cmd == 'decrypt':
    with open(BUILD_KEYS_PATH, 'r') as build_keys_file:
        enc_data = build_keys_file.read().strip()

    data = aes_decrypt(passphrase, enc_data)

    with open(BUILD_KEYS_PATH, 'w') as build_keys_file:
        build_keys_file.write(data)

    sys.exit(0)

with open(BUILD_KEYS_PATH, 'r') as build_keys_file:
    enc_data = build_keys_file.read()
    data = aes_decrypt(passphrase, enc_data)
    build_keys = json.loads(data.strip())
    github_owner = build_keys['github_owner']
    github_token = build_keys['github_token']
    gitlab_token = build_keys['gitlab_token']
    gitlab_host = build_keys['gitlab_host']

def wget(url, cwd=None, output=None):
    if output:
        args = ['wget', '-O', output, url]
    else:
        args = ['wget', url]
    subprocess.check_call(args, cwd=cwd)

def post_git_asset(release_id, file_name, file_path):
    for i in range(5):
        file_size = os.path.getsize(file_path)
        response = requests.post(
            'https://uploads.github.com/repos/%s/%s/releases/%s/assets' % (
                github_owner, REPO_NAME, release_id),
            headers={
                'Authorization': 'token %s' % github_token,
                'Content-Type': 'application/octet-stream',
                'Content-Size': str(file_size),
            },
            params={
                'name': file_name,
            },
            data=open(file_path, 'rb').read(),
        )

        if response.status_code == 201:
            return
        else:
            time.sleep(1)

    data = response.json()
    errors = data.get('errors')
    if not errors or errors[0].get('code') != 'already_exists':
        print('Failed to create asset on github')
        print(data)
        sys.exit(1)

def get_ver(version):
    day_num = (cur_date - datetime.datetime(2013, 9, 12)).days
    min_num = int(math.floor(((cur_date.hour * 60) + cur_date.minute) / 14.4))
    ver = re.findall(r'\d+', version)
    ver_str = '.'.join((ver[0], ver[1], str(day_num), str(min_num)))
    ver_str += ''.join(re.findall('[a-z]+', version))

    return ver_str

def get_int_ver(version):
    ver = re.findall(r'\d+', version)

    if 'snapshot' in version:
        pass
    elif 'alpha' in version:
        ver[-1] = str(int(ver[-1]) + 1000)
    elif 'beta' in version:
        ver[-1] = str(int(ver[-1]) + 2000)
    elif 'rc' in version:
        ver[-1] = str(int(ver[-1]) + 3000)
    else:
        ver[-1] = str(int(ver[-1]) + 4000)

    return int(''.join([x.zfill(4) for x in ver]))

def iter_packages():
    for target in BUILD_TARGETS:
        target_path = os.path.join(pacur_path, target)
        for name in os.listdir(target_path):
            if cur_version not in name:
                continue
            elif name.endswith(".pkg.tar.zst"):
                pass
            elif name.endswith(".rpm"):
                pass
            elif name.endswith(".deb"):
                pass
            else:
                continue

            path = os.path.join(target_path, name)

            yield name, path

def generate_last_modifited_etag(file_path):
    import werkzeug.http
    file_name = os.path.basename(file_path).encode(
        sys.getfilesystemencoding())
    file_mtime = datetime.datetime.utcfromtimestamp(
        os.path.getmtime(file_path))
    file_size = int(os.path.getsize(file_path))
    last_modified = werkzeug.http.http_date(file_mtime)

    return (last_modified, 'wzsdm-%d-%s-%s' % (
        time.mktime(file_mtime.timetuple()),
        file_size,
        zlib.adler32(file_name) & 0xffffffff,
    ))

with open(CONSTANTS_PATH, 'r') as constants_file:
    cur_version = re.findall('Version = "(.*?)"', constants_file.read())[0]

if cmd == 'version':
    print(get_ver(sys.argv[2]))

if cmd == 'set-version':
    new_version = get_ver(sys.argv[2])

    # Update changes
    with open(CHANGES_PATH, 'r') as changes_file:
        changes_data = changes_file.read()

    with open(CHANGES_PATH, 'w') as changes_file:
        ver_date_str = 'Version ' + new_version.replace(
            'v', '') + cur_date.strftime(' %Y-%m-%d')
        changes_file.write(changes_data.replace(
            '<%= version %>',
            '%s\n%s' % (ver_date_str, '-' * len(ver_date_str)),
        ))


    with open(CONSTANTS_PATH, 'r') as constants_file:
        constants_data = constants_file.read()

    with open(CONSTANTS_PATH, 'w') as constants_file:
        constants_file.write(re.sub(
            '(Version = ".*?")',
            'Version = "%s"' % new_version,
            constants_data,
            count=1,
        ))

    with open(CONSTANTS_PATH2, 'r') as constants_file:
        constants_data = constants_file.read()

    with open(CONSTANTS_PATH2, 'w') as constants_file:
        constants_file.write(re.sub(
            '("version": ".*?")',
            '"version": "%s"' % new_version,
            constants_data,
            count=1,
        ))

    with open(CONSTANTS_PATH3, 'r') as constants_file:
        constants_data = constants_file.read()

    with open(CONSTANTS_PATH3, 'w') as constants_file:
        constants_file.write(re.sub(
            '("version": ".*?")',
            '"version": "%s"' % new_version,
            constants_data,
            count=2,
        ))

    with open(CONSTANTS_PATH4, 'r') as constants_file:
        constants_data = constants_file.read()

    with open(CONSTANTS_PATH4, 'w') as constants_file:
        constants_file.write(re.sub(
            '(Version = ".*?")',
            'Version = "%s"' % new_version,
            constants_data,
            count=1,
        ))

    with open(CONSTANTS_PATH5, 'r') as constants_file:
        constants_data = constants_file.read()

    with open(CONSTANTS_PATH5, 'w') as constants_file:
        constants_file.write(re.sub(
            '(MyAppVersion ".*?")',
            'MyAppVersion "%s"' % new_version,
            constants_data,
            count=1,
        ))

    # Check for duplicate version
    response = requests.get(
        'https://api.github.com/repos/%s/%s/releases' % (
            github_owner, REPO_NAME),
        headers={
            'Authorization': 'token %s' % github_token,
            'Content-type': 'application/json',
        },
    )

    if response.status_code != 200:
        print('Failed to get repo releases on github')
        print(response.json())
        sys.exit(1)

    for release in response.json():
        if release['tag_name'] == new_version:
            print('Version already exists in github')
            sys.exit(1)


    # Generate changelog
    version = None
    release_body = ''
    with open(CHANGES_PATH, 'r') as changelog_file:
        for line in changelog_file.readlines()[2:]:
            line = line.strip()

            if not line or line[0] == '-':
                continue

            if line[:7] == 'Version':
                if version:
                    break
                version = line.split(' ')[1]
            elif version:
                release_body += '* %s\n' % line

    if version != new_version:
        print('New version does not exist in changes')
        sys.exit(1)

    if not release_body:
        print('Failed to generate github release body')
        sys.exit(1)
    release_body = release_body.rstrip('\n')


    subprocess.check_call(['git', 'reset', 'HEAD', '.'])
    subprocess.check_call(['git', 'add', CHANGES_PATH])
    subprocess.check_call(['git', 'add', CONSTANTS_PATH])
    subprocess.check_call(['git', 'add', CONSTANTS_PATH2])
    subprocess.check_call(['git', 'add', CONSTANTS_PATH3])
    subprocess.check_call(['git', 'add', CONSTANTS_PATH4])
    subprocess.check_call(['git', 'add', CONSTANTS_PATH5])
    subprocess.check_call(['git', 'commit', '-S', '-m', 'Create new release'])
    subprocess.check_call(['git', 'push'])


    # Create tag
    subprocess.check_call(['git', 'tag', new_version])
    subprocess.check_call(['git', 'push', '--tags'])
    time.sleep(1)


    # Create release
    response = requests.post(
        'https://api.github.com/repos/%s/%s/releases' % (
            github_owner, REPO_NAME),
        headers={
            'Authorization': 'token %s' % github_token,
            'Content-type': 'application/json',
        },
        data=json.dumps({
            'tag_name': new_version,
            'name': '%s v%s' % (RELEASE_NAME, new_version),
            'body': release_body,
            'prerelease': False,
            'target_commitish': 'master',
        }),
    )

    if response.status_code != 201:
        print('Failed to create release on github')
        print(response.json())
        sys.exit(1)

    subprocess.check_call(['git', 'pull'])
    subprocess.check_call(['git', 'push', '--tags'])
    time.sleep(6)


    # Create gitlab release
    response = requests.post(
        ('https://%s/api/v4/projects' +
         '/%s%%2F%s/releases') % (
            gitlab_host, github_owner, REPO_NAME),
        headers={
            'Private-Token': gitlab_token,
            'Content-type': 'application/json',
        },
        data=json.dumps({
            'tag_name': new_version,
            'name': '%s v%s' % (REPO_NAME, new_version),
            'description': release_body,
        }),
    )

    if response.status_code != 201:
        print('Failed to create release on gitlab')
        print(response.json())
        sys.exit(1)


if cmd == 'build' or cmd == 'build-test' or cmd == 'build-upload':
    if cmd == 'build' or cmd == 'build-upload':
        pacur_path = STABLE_PACUR_PATH
    else:
        pacur_path = TEST_PACUR_PATH


    # Get sha256 sum
    archive_name = '%s.tar.gz' % cur_version
    archive_path = os.path.join(os.path.sep, 'tmp', archive_name)
    if os.path.isfile(archive_path):
        os.remove(archive_path)
    wget('https://github.com/%s/%s/archive/%s' % (
        github_owner, REPO_NAME, archive_name),
        output=archive_name,
        cwd=os.path.join(os.path.sep, 'tmp'),
    )
    archive_sha256_sum = subprocess.check_output(
        ['sha256sum', archive_path]).split()[0]
    os.remove(archive_path)


    for target in BUILD_TARGETS:
        pkgbuild_path = os.path.join(pacur_path, target, 'PKGBUILD')

        with open(pkgbuild_path, 'r') as pkgbuild_file:
            pkgbuild_data = re.sub(
                'pkgver="(.*)"',
                'pkgver="%s"' % cur_version,
                pkgbuild_file.read(),
                count=1,
            )
            pkgbuild_data = re.sub(
                '"[a-f0-9]{64}"',
                '"%s"' % archive_sha256_sum.decode('utf-8'),
                pkgbuild_data,
                count=1,
            )

        with open(pkgbuild_path, 'w') as pkgbuild_file:
            pkgbuild_file.write(pkgbuild_data)

    for build_target in BUILD_TARGETS:
        subprocess.check_call(
            ['sudo', 'pacur', 'project', 'build', build_target],
            cwd=pacur_path,
        )

if cmd == 'upload' or cmd == 'upload-test' or cmd == 'build-upload':
    pacur_path = TEST_PACUR_PATH if is_snapshot else STABLE_PACUR_PATH

    # Get release id
    release_id = None
    response = requests.get(
        'https://api.github.com/repos/%s/%s/releases' % (
            github_owner, REPO_NAME),
        headers={
            'Authorization': 'token %s' % github_token,
            'Content-type': 'application/json',
        },
    )

    for release in response.json():
        if release['tag_name'] == cur_version:
            release_id = release['id']

    if not release_id:
        print('Version does not exists in github')
        sys.exit(1)


    subprocess.check_call(
        ['sudo', 'pacur', 'project', 'repo'],
        cwd=pacur_path,
    )

    for name, path in iter_packages():
        post_git_asset(release_id, name, path)

    subprocess.check_call([
        'sh',
        'upload-unstable.sh',
    ], cwd=pacur_path)
