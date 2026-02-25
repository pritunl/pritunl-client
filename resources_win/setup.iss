#define MyAppName "Pritunl"
#define MyAppVersion "1.3.4466.51"
#define MyAppPublisher "Pritunl"
#define MyAppURL "https://pritunl.com/"
#define MyAppExeName "pritunl.exe"

[Setup]
AppId={#MyAppName}
AppName={#MyAppName}
AppVersion={#MyAppVersion}
VersionInfoVersion={#MyAppVersion}
AppPublisher={#MyAppPublisher}
AppPublisherURL={#MyAppURL}
AppSupportURL={#MyAppURL}
AppUpdatesURL={#MyAppURL}
DefaultDirName={pf}\{#MyAppName}
DefaultGroupName={#MyAppName}
PrivilegesRequired=admin
DisableProgramGroupPage=yes
OutputDir=..\build\
OutputBaseFilename={#MyAppName}
LicenseFile=license.txt
SetupIconFile=..\client\www\img\logo.ico
CloseApplications=force
UninstallDisplayName=Pritunl Client
UninstallDisplayIcon={app}\{#MyAppExeName}
Compression=lzma
SolidCompression=yes
SignTool=signtool

[Languages]
Name: "english"; MessagesFile: "compiler:Default.isl"

[Tasks]
Name: "desktopicon"; Description: "{cm:CreateDesktopIcon}"; GroupDescription: "{cm:AdditionalIcons}"; Flags: checkedonce

[Files]
Source: "..\build\win\pritunl-win32-x64\*"; DestDir: "{app}"; Flags: ignoreversion recursesubdirs createallsubdirs; Check: not IsArm64
Source: "..\build\win\pritunl-win32-arm64\*"; DestDir: "{app}"; Flags: ignoreversion recursesubdirs createallsubdirs; Check: IsArm64
Source: "..\tuntap_win\tuntap_amd64\*"; DestDir: "{app}\tuntap"; Flags: ignoreversion recursesubdirs createallsubdirs; Check: not IsArm64
Source: "..\tuntap_win\tuntap10_amd64\*"; DestDir: "{app}\tuntap10"; Flags: ignoreversion recursesubdirs createallsubdirs; Check: not IsArm64
Source: "..\openvpn_win\openvpn_amd64\*"; DestDir: "{app}\openvpn"; Flags: ignoreversion recursesubdirs createallsubdirs; Check: not IsArm64
Source: "..\wireguard_win\wireguard_amd64\*"; DestDir: "{app}\wireguard"; Flags: ignoreversion recursesubdirs createallsubdirs; Check: not IsArm64
Source: "..\tuntap_win\tuntap_arm64\*"; DestDir: "{app}\tuntap"; Flags: ignoreversion recursesubdirs createallsubdirs; Check: IsArm64
Source: "..\tuntap_win\tuntap10_arm64\*"; DestDir: "{app}\tuntap10"; Flags: ignoreversion recursesubdirs createallsubdirs; Check: IsArm64
Source: "..\openvpn_win\openvpn_arm64\*"; DestDir: "{app}\openvpn"; Flags: ignoreversion recursesubdirs createallsubdirs; Check: IsArm64
Source: "..\wireguard_win\wireguard_arm64\*"; DestDir: "{app}\wireguard"; Flags: ignoreversion recursesubdirs createallsubdirs; Check: IsArm64
Source: "..\service\service_amd64.exe"; DestDir: "{app}"; DestName: "pritunl-service.exe"; Flags: ignoreversion; Check: not IsArm64
Source: "..\cli\cli_amd64.exe"; DestDir: "{app}"; DestName: "pritunl-client.exe"; Flags: ignoreversion; Check: not IsArm64
Source: "..\service\service_arm64.exe"; DestDir: "{app}"; DestName: "pritunl-service.exe"; Flags: ignoreversion; Check: IsArm64
Source: "..\cli\cli_arm64.exe"; DestDir: "{app}"; DestName: "pritunl-client.exe"; Flags: ignoreversion; Check: IsArm64

[Code]
procedure StopApplication();
var ResultCode: Integer;
begin
    Exec('sc.exe', 'stop pritunl', '', SW_HIDE, ewWaitUntilTerminated, ResultCode);
    Sleep(3000);
    Exec(
        'powershell.exe',
        '-NoProfile -ExecutionPolicy Bypass -Command "Get-CimInstance Win32_Process | Where-Object { $_.Name -eq ''pritunl.exe'' -and $_.ExecutablePath -like ''*\Program Files (x86)\Pritunl\pritunl.exe'' } | ForEach-Object { Stop-Process -Id $_.ProcessId -Force }"',
        '',
        SW_HIDE,
        ewWaitUntilTerminated,
        ResultCode
    );
    Sleep(500);
end;
function PrepareToInstall(var NeedsRestart: Boolean): String;
begin
    StopApplication();
    Result := '';
end;
function InitializeUninstall(): Boolean;
begin
    StopApplication();
    Result := True;
end;

[Icons]
Name: "{group}\{#MyAppName}"; Filename: "{app}\{#MyAppExeName}"
Name: "{group}\{cm:UninstallProgram,{#MyAppName}}"; Filename: "{uninstallexe}"
Name: "{commonstartup}\{#MyAppName}"; Filename: "{app}\{#MyAppExeName}"; Parameters: "--no-main"
Name: "{commondesktop}\{#MyAppName}"; Filename: "{app}\{#MyAppExeName}"; Tasks: desktopicon

[InstallDelete]
Type: filesandordirs; Name: "{app}"

[Run]
Filename: "{app}\pritunl-service.exe"; Parameters: "-install"; Flags: runhidden; StatusMsg: "Configuring Pritunl..."

[UninstallRun]
Filename: "{app}\pritunl-service.exe"; Parameters: "-uninstall"; Flags: runhidden

[UninstallDelete]
Type: filesandordirs; Name: "{app}"
Type: filesandordirs; Name: "C:\ProgramData\{#MyAppName}"
Type: filesandordirs; Name: "{userappdata}\pritunl"
