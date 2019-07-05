from pritunl_client.constants import *
from pritunl_client.exceptions import *
from pritunl_client import profile
from pritunl_client import utils
from pritunl_client import logger

import BaseHTTPServer
import SocketServer
import threading
import json
import httplib
import urllib
from socket import error as socket_error

auth_token = utils.generate_secret()

class ThreadingHTTPServer(SocketServer.ThreadingMixIn,
        BaseHTTPServer.HTTPServer):
    pass

class Request(BaseHTTPServer.BaseHTTPRequestHandler):
    def prfl_lookup(self):
        vpn_list_url = 'http://localhost:9797/list'
        vpn_table = urllib.urlopen(vpn_list_url)
        vpn_headers = vpn_table.readline().strip().split()
        vpn_info = []
        for row in vpn_table:
            row = row.strip().split()
            vpn_data = dict(zip(vpn_headers, row))
            vpn_info.append(vpn_data)
        for vpn_data in vpn_info:
            if self.args[2] == vpn_data['NAME']:
                self.args[2] = vpn_data['PROFILE_ID']
        return self.args[2]

    def send_text_response(self, text, status_code=200):
        self.send_response(status_code)
        self.send_header('Content-type', 'text/plain')
        self.end_headers()

        self.wfile.write(text)
        self.wfile.close()

    def do_PUT(self):
        try:
            if self.headers.get('Auth-Token') != auth_token or \
                    self.headers.get('User-Agent') != 'pritunl' or \
                    self.headers.get('Origin') or \
                    self.headers.get('Referer'):
                self.send_response(401)
                return

            self.args = self.path.split('/')

            data_len = int(self.headers.get('Content-Length', 0))
            data = self.rfile.read(data_len)

            if data:
                self.data = json.loads(data)
            else:
                self.data = {}

            if self.args[1] == 'remove':
                self.do_remove()
            elif self.args[1] == 'start':
                self.do_start()
            elif self.args[1] == 'stop':
                self.do_stop()
            elif self.args[1] == 'enable':
                self.do_enable()
            elif self.args[1] == 'disable':
                self.do_disable()
            else:
                self.send_response(404)
        except Exception as exception:
            self.send_text_response(str(exception), 500)

    def do_POST(self):
        try:
            if self.headers.get('Auth-Token') != auth_token or \
                    self.headers.get('User-Agent') != 'pritunl' or \
                    self.headers.get('Origin') or \
                    self.headers.get('Referer'):
                self.send_response(401)
                return

            self.args = self.path.split('/')

            data_len = int(self.headers.get('Content-Length', 0))
            data = self.rfile.read(data_len)

            if data:
                self.data = json.loads(data)
            else:
                self.data = {}

            if self.args[1] == 'import':
                self.do_import()
            else:
                self.send_response(404)
        except Exception as exception:
            self.send_text_response(str(exception), 500)

    def do_DELETE(self):
        try:
            if self.headers.get('Auth-Token') != auth_token or \
                    self.headers.get('User-Agent') != 'pritunl' or \
                    self.headers.get('Origin') or \
                    self.headers.get('Referer'):
                self.send_response(401)
                return

            self.args = self.path.split('/')

            data_len = int(self.headers.get('Content-Length', 0))
            data = self.rfile.read(data_len)

            if data:
                self.data = json.loads(data)
            else:
                self.data = {}

            if self.args[1] == 'remove':
                self.do_remove()
            else:
                self.send_response(404)
        except Exception as exception:
            self.send_text_response(str(exception), 500)

    def do_GET(self):
        try:
            self.args = self.path.split('/')

            if self.args[1] == 'token':
                self.do_token()
            elif self.args[1] == 'list':
                if self.headers.get('Auth-Token') != auth_token or \
                        self.headers.get('User-Agent') != 'pritunl' or \
                        self.headers.get('Origin') or \
                        self.headers.get('Referer'):
                    self.send_response(401)
                    return

                self.do_list()
            else:
                self.send_response(404)
        except Exception as exception:
            self.send_text_response(str(exception), 500)

    def do_token(self):
        self.send_text_response(auth_token, 200)

    def do_import(self):
        profile_path = self.data.get('profile_path')
        profile_uri = self.data.get('profile_uri')

        if profile_path:
            profile.import_file(profile_path)
        elif profile_uri:
            try:
                profile.import_uri(profile_uri)
            except httplib.HTTPException:
                self.send_text_response('Unable to retrieve profile or ' +
                    'profile path does not exist', 500)
        else:
            raise ValueError('Must have profile_path or profile_uri')

        self.send_response(200)

    def do_remove(self):
        self.args[2] = self.prfl_lookup()
        prfl = profile.Profile.get_profile(self.args[2])
        if prfl:
            prfl.delete()

        self.send_response(200)

    def do_start(self):
        passwd = self.data.get('passwd')
        evt = threading.Event()

        def status_callback():
            pass

        def connect_callback():
            evt.set()
            pass

        self.args[2] = self.prfl_lookup()
        prfl = profile.Profile.get_profile(self.args[2])
        if not prfl:
            self.send_text_response('Profile not found', 404)
            return

        prfl.sync_conf()

        if prfl.auth_type and not passwd:
            self.send_text_response('Password required', 400)
            return

        if not prfl.start(status_callback, connect_callback, passwd=passwd):
            self.send_text_response('Profile has already been started', 500)
            return

        evt.wait(CONNECT_TIMEOUT + 5)

        if prfl.status in ERROR_STATES:
            error_msg = {
                ERROR: 'An error occurred while connecting to server',
                AUTH_ERROR: 'Failed to authenticate with server',
                TIMEOUT_ERROR: 'Server connection timed out',
            }[prfl.status]

            self.send_text_response(
                error_msg, 500)
            return

        self.send_response(200)

    def do_stop(self):
        self.args[2] = self.prfl_lookup()
        prfl = profile.Profile.get_profile(self.args[2])
        if not prfl:
            self.send_text_response('Profile not found', 404)
            return
        prfl.stop()

        self.send_response(200)

    def do_enable(self):
        self.args[2] = self.prfl_lookup()
        prfl = profile.Profile.get_profile(self.args[2])
        if not prfl:
            self.send_text_response('Profile not found', 404)
            return
        prfl.set_autostart(True)

        self.send_response(200)

    def do_disable(self):
        self.args[2] = self.prfl_lookup()
        prfl = profile.Profile.get_profile(self.args[2])
        if not prfl:
            self.send_text_response('Profile not found', 404)
            return
        prfl.set_autostart(False)

        self.send_response(200)

    def do_list(self):
        output = '%s %s %s %s\n' % (
            'PROFILE_ID'.ljust(33, ' '),
            'ENABLED'.ljust(8, ' '),
            'STATE'.ljust(13, ' '),
            'NAME',
        )

        for prfl in profile.Profile.iter_profiles():
            output += '%s %s %s %s\n' % (
                prfl.id.ljust(33, ' '),
                str(prfl.autostart).lower().ljust(8, ' '),
                prfl.status.ljust(13, ' '),
                prfl.name,
            )

        self.send_text_response(output.rstrip('\n'))

class ShellApp(object):
    def __init__(self):
        self.start_server()

    def autostart(self):
        def status_callback():
            pass

        for prfl in profile.Profile.iter_profiles():
            if not prfl.autostart:
                continue

            logger.info('Auto starting profile...', 'shell',
                profile_id=prfl.id,
                profile_name=prfl.name,
            )

            def connect_callback():
                if prfl.status in ERROR_STATES:
                    logger.error('Failed to autostart server', 'shell',
                        profile_id=prfl.id,
                        profile_name=prfl.name,
                        profile_status=prfl.status,
                    )

            threading.Thread(target=prfl.start_autostart,
                args=(status_callback, connect_callback)).start()

    def start_server(self):
        try:
            server = ThreadingHTTPServer(
                ('127.0.0.1', 9797),
                Request,
            )
            logger.info('Starting pritunl-client daemon...', 'shell')

            self.autostart()

            server.serve_forever()
        except socket_error:
            logger.info('Address already in use. Make sure that port 9797 is available.', 'shell')
