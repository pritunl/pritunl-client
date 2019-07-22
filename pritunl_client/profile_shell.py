from constants import *
from exceptions import *
from pritunl_client import profile

import os
import time
import signal

class ProfileShell(profile.Profile):
    def _start(self, status_callback, connect_callback, passwd):
        def on_exit(return_code):
            if self.status in ACTIVE_STATES:
                self._set_status(ERROR)

        args = ['openvpn', '--config', self.path]

        if passwd:
            args.append('--auth-user-pass')
            args.append(self.passwd_path)

            with open(self.passwd_path, 'w') as passwd_file:
                os.chmod(self.passwd_path, 0600)
                passwd_file.write('pritunl_client\n')
                passwd_file.write('%s\n' % passwd)

        systemd_resolve = False
        try:
            subprocess.check_call(['which', 'systemd-resolve'])
            with open('/etc/resolv.conf', 'r') as resolv_file:
                data = resolv_file.read()
                if 'systemd-resolved' in data or '127.0.0.53' in data:
                    systemd_resolve = True
        except:
            pass
        if systemd_resolve:
            script_path = os.path.join(SHARE_DIR, 'update-systemd-resolved.sh')
        else:
            script_path = os.path.join(SHARE_DIR, 'update-resolv-conf.sh')

        args.extend(['--script-security', '2'])
        args.append('--up-restart')
        args.extend(['--up', script_path])
        args.extend(['--down', script_path])

        self._run_ovpn(status_callback, connect_callback, args, on_exit, False)

    def _start_autostart(self, status_callback, connect_callback):
        self._start(status_callback, connect_callback, None)

    def _stop(self, silent):
        data = profile._connections.get(self.id)
        if data:
            process = data.get('process')
            data['process'] = None
            if process:
                process.terminate()
                for i in xrange(int(5 / 0.1)):
                    time.sleep(0.1)
                    if process.poll() is not None:
                        break
                    process.terminate()

                for i in xrange(int(5 / 0.1)):
                    time.sleep(0.1)
                    if process.poll() is not None:
                        break
                    process.kill()
        if not silent:
            self._set_status(ENDED)
        self.pid = None
        self.commit()

    def _kill_pid(self, pid):
        for i in xrange(2):
            try:
                os.kill(pid, signal.SIGTERM)
            except OSError:
                pass
