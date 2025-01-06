; Basic NSIS based KmAgent Installation Script
; Written for NSIS 3.0 or higher

BrandingText "KloudMate Technologies Inc. All rights Reserved."
!define APPNAME "KmAgent"
!define COMPANYNAME "Kloudmate"
!define DESCRIPTION "A short description of Km Agent"
!define VERSIONMAJOR 1
!define VERSIONMINOR 0
!define VERSIONBUILD 0
!define YAML_FILENAME "agent-config.yaml"
!define YAML_INSTALL_DIR "$PROFILE\.kloudmate"
!define SERVICE_NAME "KmAgent"
!define SERVICE_DISPLAY_NAME "Kloudmate Agent"
!define SERVICE_DESCRIPTION "Kloudmate Agent Description"

; Define installer attributes
Name "${APPNAME}"
OutFile "km-agent_yaml_with_service_Setup.exe"
InstallDir "$PROGRAMFILES\${COMPANYNAME}\${APPNAME}"
InstallDirRegKey HKLM "Software\${COMPANYNAME}\${APPNAME}" "Install_Dir"

; Request admin privileges
RequestExecutionLevel admin

; Modern UI
!include "MUI2.nsh"

; UI Settings
!define MUI_ABORTWARNING
!define MUI_ICON "${NSISDIR}\Contrib\Graphics\Icons\modern-install.ico"
!define MUI_UNICON "${NSISDIR}\Contrib\Graphics\Icons\modern-uninstall.ico"

; Custom finish page settings
!define MUI_FINISHPAGE_TITLE "Installation Complete"
!define MUI_FINISHPAGE_TEXT_LARGE
!define MUI_FINISHPAGE_TEXT "Setup has finished installing ${APPNAME} on your computer.$\n$\nYAML Configuration File Location:$\n${YAML_INSTALL_DIR}\${YAML_FILENAME}$\n$\nThe service has been installed and started.$\n$\nClick Finish to close Setup."
!define MUI_FINISHPAGE_RUN ""  ; Disabled since it's a service

; Pages
!insertmacro MUI_PAGE_WELCOME
!insertmacro MUI_PAGE_LICENSE "license.txt"
!insertmacro MUI_PAGE_DIRECTORY
!insertmacro MUI_PAGE_INSTFILES
!insertmacro MUI_PAGE_FINISH

; Uninstaller pages
!insertmacro MUI_UNPAGE_CONFIRM
!insertmacro MUI_UNPAGE_INSTFILES

; Language
!insertmacro MUI_LANGUAGE "English"

; Service installation and control functions
Function InstallService
    ; Install the service
    nsExec::ExecToLog '"$INSTDIR\km-agent-ci.exe" install'
    Pop $0
    ${If} $0 != 0
        MessageBox MB_OK|MB_ICONSTOP "Failed to install service. Error code: $0"
    ${EndIf}
FunctionEnd

Function StartService
    ; Start the service
    nsExec::ExecToLog 'net start ${SERVICE_NAME}'
    Pop $0
    ${If} $0 != 0
        MessageBox MB_OK|MB_ICONSTOP "Failed to start service. Error code: $0"
    ${EndIf}
FunctionEnd

Function StopService
    ; Stop the service
    nsExec::ExecToLog 'net stop ${SERVICE_NAME}'
    Pop $0
FunctionEnd

Function un.StopAndRemoveService
    ; Stop and remove the service during uninstall
    nsExec::ExecToLog 'net stop ${SERVICE_NAME}'
    nsExec::ExecToLog '"$INSTDIR\km-agent-ci.exe" uninstall'
FunctionEnd

; Installer sections
Section "MainApplication" SecMain
    SectionIn RO  ; Read-only, always installed
    SetOutPath "$INSTDIR"
    
    ; Add your main application files here
    File "km-agent-ci.exe"
    File "readme.txt"
    
    ; Create Start Menu shortcuts
    CreateDirectory "$SMPROGRAMS\${COMPANYNAME}"
    CreateShortcut "$SMPROGRAMS\${COMPANYNAME}\Uninstall ${APPNAME}.lnk" "$INSTDIR\uninstall.exe"
    
    ; Write registry keys for uninstall
    WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${COMPANYNAME} ${APPNAME}" "DisplayName" "${APPNAME}"
    WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${COMPANYNAME} ${APPNAME}" "UninstallString" "$\"$INSTDIR\uninstall.exe$\""
    WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${COMPANYNAME} ${APPNAME}" "QuietUninstallString" "$\"$INSTDIR\uninstall.exe$\" /S"
    WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${COMPANYNAME} ${APPNAME}" "InstallLocation" "$\"$INSTDIR$\""
    WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${COMPANYNAME} ${APPNAME}" "DisplayIcon" "$\"$INSTDIR\km-agent-ci.exe$\""
    WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${COMPANYNAME} ${APPNAME}" "Publisher" "${COMPANYNAME}"
    WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${COMPANYNAME} ${APPNAME}" "DisplayVersion" "${VERSIONMAJOR}.${VERSIONMINOR}.${VERSIONBUILD}"
    
    ; Store YAML file location in registry for the application to read
    WriteRegStr HKLM "Software\${COMPANYNAME}\${APPNAME}" "YAMLLocation" "${YAML_INSTALL_DIR}\${YAML_FILENAME}"
    
    ; Create uninstaller
    WriteUninstaller "$INSTDIR\uninstall.exe"
    
    ; Install and start the service
    Call InstallService
    Call StartService
SectionEnd

Section "YAMLConfiguration" SecYAML
    ; Create YAML directory if it doesn't exist
    CreateDirectory "${YAML_INSTALL_DIR}"
    
    ; Set output path to YAML directory and copy the YAML file
    SetOutPath "${YAML_INSTALL_DIR}"
    File "${YAML_FILENAME}"
    
    ; Store YAML location for uninstaller
    WriteRegStr HKLM "Software\${COMPANYNAME}\${APPNAME}" "YAMLPath" "${YAML_INSTALL_DIR}"
SectionEnd

; Uninstaller section
Section "Uninstall"
    ; Stop and remove the service
    Call un.StopAndRemoveService
    
    ; Remove Start Menu shortcuts
    Delete "$SMPROGRAMS\${COMPANYNAME}\Uninstall ${APPNAME}.lnk"
    RMDir "$SMPROGRAMS\${COMPANYNAME}"
    
    ; Remove YAML file
    Delete "${YAML_INSTALL_DIR}\${YAML_FILENAME}"
    RMDir "${YAML_INSTALL_DIR}"  ; Remove directory if empty
    
    ; Remove main application files
    Delete "$INSTDIR\km-agent-ci.exe"
    Delete "$INSTDIR\readme.txt"
    Delete "$INSTDIR\uninstall.exe"
    
    ; Remove install directory
    RMDir "$INSTDIR"
    
    ; Remove registry keys
    DeleteRegKey HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${COMPANYNAME} ${APPNAME}"
    DeleteRegKey HKLM "Software\${COMPANYNAME}\${APPNAME}"
SectionEnd