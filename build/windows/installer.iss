[Setup]
AppName=MyApp
AppVersion=1.0
DefaultDirName={pf}\MyApp
DefaultGroupName=MyApp
OutputDir=dist\windows
OutputBaseFilename=MyAppInstaller
Compression=lzma
SolidCompression=yes

[Files]
Source: "dist\windows\myapp.exe"; DestDir: "{app}"; Flags: ignoreversion

[Dirs]
Name: "{commonappdata}\MyApp"

[Code]
var
  ApiInput: string;

function InitializeSetup(): Boolean;
begin
  // Ask user for KM_API
  Result := InputQuery('Configuration', 'Enter your KM_API endpoint:', ApiInput);
end;

procedure CurStepChanged(CurStep: TSetupStep);
var
  ConfigFile: string;
  ConfigContent: string;
begin
  if CurStep = ssPostInstall then begin
    ConfigFile := ExpandConstant('{commonappdata}\MyApp\config.yaml');
    ConfigContent :=
      'api: "' + ApiInput + '"' + #13#10 +
      'debug: false' + #13#10;

    SaveStringToFile(ConfigFile, ConfigContent, False);
  end;
end;

[Run]
Filename: "{app}\myapp.exe"; Parameters: "install --config ""{commonappdata}\MyApp\config.yaml"""; Flags: runhidden
