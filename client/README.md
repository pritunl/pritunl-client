### pritunl-client-www

```
npm install
```

#### locked packages

```
npx webpack
```

### development

```
cd service
sudo ./service

cd client
sed -i 's|dist/static|dist-dev/static|g' package.json
npx tsc --watch
npx webpack-cli --config webpack.dev.config --progress --color --watch
npx webpack-cli --config webpack-main.dev.config --progress --color --watch
npx electron . --dev-tools
```

#### production

```
sh build.sh
```

### clean

```
rm -rf app/*.js*
rm -rf app/**/*.js*
```

### internal

```
# desktop
rsync --human-readable --archive --xattrs --progress --delete --exclude "/node_modules/*" --exclude "/jspm_packages/*" --exclude "app/*.js" --exclude "app/*.js.map" --exclude "app/**/*.js" --exclude "app/**/*.js.map" /home/cloud/go/src/github.com/pritunl/pritunl-client/client/ $NPM_SERVER:/home/cloud/pritunl-client-www/

# npm-server
cd /home/cloud/pritunl-cloud-www/
rm -rf node_modules
npm install

# desktop
scp $NPM_SERVER:/home/cloud/pritunl-client-www/package.json /home/cloud/go/src/github.com/pritunl/pritunl-client/client/package.json
scp $NPM_SERVER:/home/cloud/pritunl-client-www/package-lock.json /home/cloud/go/src/github.com/pritunl/pritunl-client/client/package-lock.json
rsync --human-readable --archive --xattrs --progress --delete $NPM_SERVER:/home/cloud/pritunl-client-www/node_modules/ /home/cloud/go/src/github.com/pritunl/pritunl-client/client/node_modules/
rsync --human-readable --archive --xattrs --progress --delete --exclude "/node_modules/*" --exclude "/jspm_packages/*" --exclude "app/*.js" --exclude "app/*.js.map" --exclude "app/**/*.js" --exclude "app/**/*.js.map" /home/cloud/go/src/github.com/pritunl/pritunl-client/client/ $NPM_SERVER:/home/cloud/pritunl-client-www/

# npm-server
sh build.sh

# desktop
rsync --human-readable --archive --xattrs --progress --delete $NPM_SERVER:/home/cloud/pritunl-client-www/dist/ /home/cloud/go/src/github.com/pritunl/pritunl-client/client/dist/
rsync --human-readable --archive --xattrs --progress --delete $NPM_SERVER:/home/cloud/pritunl-client-www/dist-dev/ /home/cloud/go/src/github.com/pritunl/pritunl-client/client/dist-dev/
```
