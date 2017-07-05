from constants import *
from exceptions import *
from pritunl_client import profile
from pritunl_client import utils

import os
import time
import subprocess
import hashlib

class ProfileLinux(profile.Profile):
    def _get_profile_hash(self):
        if self.encrypted and not self.decrypted_data:
            return

        if os.path.exists(self.path):
            return hashlib.sha512(self.get_vpn_conf()).hexdigest()

    def _get_profile_hash_path(self):
        profile_hash = self._get_profile_hash()
        if profile_hash:
            return os.path.join(os.path.abspath(os.sep),
                'etc', 'pritunl_client', profile_hash)

    def _start(self, status_callback, connect_callback, auth_token,
            username_passwd, mode=START):
        if self.autostart or mode == AUTOSTART:
            if not os.path.exists(self._get_profile_hash_path()):
                self.set_autostart(False)
                if mode == AUTOSTART:
                    return
            else:
                mode = AUTOSTART

        def on_exit(return_code):
            # Canceled
            if return_code == 126:
                self._set_status(ENDED)
            else:
                if self.status in ACTIVE_STATES:
                    self._set_status(ERROR)

        args = ['pkexec', '/usr/bin/pritunl-client-pk-%s' % mode]

        env = {'VPN_CONF': self.get_vpn_conf()}
        if auth_token or username_passwd:
            if auth_token:
                auth_token += '<%=AUTH_TOKEN=%>'
                if username_passwd:
                    username_passwd = auth_token + username_passwd
                else:
                    username_passwd = auth_token

            env['VPN_PASSWORD'] = username_passwd[0]
            env['VPN_USERNAME'] = username_passwd[1]

        self._run_ovpn(status_callback, connect_callback,
            args, on_exit, True, env=env)

    def _start_autostart(self, status_callback, connect_callback, auth_token):
        self._start(status_callback, connect_callback, auth_token,
            None, AUTOSTART)

    def _stop(self, silent):
        data = profile._connections.get(self.id)
        if data:
            process = data.get('process')
            data['process'] = None
            if process and not process.poll():
                for _ in range(250):
                    stop_process = subprocess.Popen(['pkexec',
                        '/usr/bin/pritunl-client-pk-stop', str(process.pid)])
                    stop_process.wait()

                    # Canceled
                    if stop_process.returncode == 126:
                        return
                    # Random error, retry
                    elif stop_process.returncode == -15:
                        time.sleep(0.05)
                        continue
                    elif stop_process.returncode != 0:
                        raise ProcessCallError(
                            'Pritunl polkit process returned error %s.' % (
                                stop_process.returncode))
                    else:
                        break
        if not silent:
            self._set_status(ENDED)
        self.pid = None
        self.commit()

    def _set_profile_autostart(self):
        for _ in xrange(250):
            process = subprocess.Popen([
                'pkexec',
                '/usr/bin/pritunl-client-pk-set-autostart',
                utils.write_env({'VPN_CONF': self.get_vpn_conf()}),
            ])
            process.wait()

            # Canceled
            if process.returncode == 126:
                return False
            # Random error, retry
            elif process.returncode == -15:
                time.sleep(0.05)
                continue
            elif process.returncode != 0:
                raise ProcessCallError(
                    'Pritunl polkit process returned error %s.' % (
                        process.returncode))
            else:
                break
        return True

    def _clear_profile_autostart(self):
        for _ in xrange(250):
            process = subprocess.Popen(['pkexec',
                '/usr/bin/pritunl-client-pk-clear-autostart',
                self._get_profile_hash()])
            process.wait()

            # Canceled
            if process.returncode == 126:
                return False
            # Random error, retry
            elif process.returncode == -15:
                time.sleep(0.05)
                continue
            elif process.returncode != 0:
                raise ProcessCallError(
                    'Pritunl polkit process returned error %s.' % (
                        process.returncode))
            else:
                break
        return True

    def _kill_pid(self, pid):
        for _ in xrange(250):
            process = subprocess.Popen(['pkexec',
                '/usr/bin/pritunl-client-pk-stop', str(pid)],
                stdout=subprocess.PIPE, stderr=subprocess.PIPE)
            process.wait()

            # Random error, retry
            if process.returncode == -15:
                time.sleep(0.05)
                continue
            else:
                break

    def commit(self):
        profile_hash_path = self._get_profile_hash_path()
        if profile_hash_path and \
                os.path.exists(profile_hash_path) != self.autostart:
            if self.autostart:
                if not self._set_profile_autostart():
                    return
            else:
                if not self._clear_profile_autostart():
                    return
        profile.Profile.commit(self)

    def delete(self):
        profile_hash_path = self._get_profile_hash_path()
        if profile_hash_path and os.path.exists(profile_hash_path):
            if not self._clear_profile_autostart():
                return
        profile.Profile.delete(self)
