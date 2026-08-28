const packager = require('@electron/packager');
const path = require("path");
const fs = require("fs");
const fuses = require("@electron/fuses");

let entitlementsPath = path.resolve(__dirname, '..',
  'resources_macos', 'entitlements.plist');

async function packageApp() {
  try {
    const appPaths = await packager({
      dir: './',
      name: 'Pritunl',
      platform: 'darwin',
      arch: 'universal',
      icon: './www/img/pritunl.icns',
      darwinDarkModeSupport: true,
      extraResource: [
        '../build/resources/pritunl-client',
        '../build/resources/Pritunl Device Authentication',
      ],
      osxUniversal: {
        x64ArchFiles: '*'
      },
      osxSign: {
        hardenedRuntime: true,
        // TODO
        // optionsForFile: (filePath) => {
        //   return {
        //     entitlements: entitlementsPath,
        //     hardenedRuntime: true,
        //   }
        // },
        identity: 'Developer ID Application: Pritunl, Inc. (U22BLATN63)'
      },
      osxNotarize: {
        keychainProfile: 'Pritunl',
        tool: 'notarytool'
      },
      asar: true,
      out: '../build/macos/Applications',
      gatekeeperAssess: false,
      afterCopyExtraResources: [
        async (buildPath, electronVersion,
            platform, arch, callback) => {
          console.log(`Packaging app for ${platform}-${arch} ` +
            `using Electron ${electronVersion} in ${buildPath}`);

          let electronPath = path.resolve(buildPath,
            'Pritunl.app/Contents/MacOS/Electron');

          console.log(`Flip fuses in ${electronPath}`);

          await fuses.flipFuses(electronPath, {
            version: fuses.FuseVersion.V1,
            [fuses.FuseV1Options.RunAsNode]: false,
            [fuses.FuseV1Options.EnableNodeOptionsEnvironmentVariable]: false,
            [fuses.FuseV1Options.EnableNodeCliInspectArguments]: false,
            [fuses.FuseV1Options.EnableEmbeddedAsarIntegrityValidation]: true,
            [fuses.FuseV1Options.OnlyLoadAppFromAsar]: true,
          });

          let daemonsPath = path.resolve(buildPath,
            'Pritunl.app/Contents/Library/LaunchDaemons');

          console.log(`Copy service daemon plist to ${daemonsPath}`);

          fs.mkdirSync(daemonsPath, {recursive: true});
          fs.copyFileSync(
            path.resolve(__dirname, '..', 'service_macos',
              'com.pritunl.service.sma.plist'),
            path.resolve(daemonsPath, 'com.pritunl.service.plist'),
          );

          let helperPath = path.resolve(buildPath,
            'Pritunl.app/Contents/MacOS/pritunl-service-helper');

          console.log(`Copy service helper to ${helperPath}`);

          fs.copyFileSync(
            path.resolve(__dirname, '..', 'build', 'resources',
              'pritunl-service-helper'),
            helperPath,
          );
          fs.chmodSync(helperPath, 0o755);

          callback();
        }
      ]
    });

    console.log(`Successfully packaged app at ${appPaths}`);
  } catch (err) {
    console.error('Error packaging app:', err);
  }
}

packageApp();
