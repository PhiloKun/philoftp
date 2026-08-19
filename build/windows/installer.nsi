; PhiloFTP Windows 安装程序
; 使用 NSIS + Modern UI 2，支持标准安装流程（选目录/快捷方式/卸载）

Unicode True
RequestExecutionLevel user
SetCompressor /SOLID lzma

!include "MUI2.nsh"
!include "FileFunc.nsh"

; ---------- 常量 ----------
; __VERSION__ / __VERSION4__ / __EXE_NAME__ 由 Makefile 构建时替换为实际版本与产物文件名
!define APP_NAME "PhiloFTP"
!define APP_VERSION "__VERSION__"
!define APP_PUBLISHER "PhiloKun"
!define APP_WEB "https://gitee.com/PhiloKun/philoftp"
!define REG_UNINSTALL "Software\Microsoft\Windows\CurrentVersion\Uninstall\PhiloFTP"

Name "${APP_NAME}"
OutFile "PhiloFTP-Setup.exe"
; 默认安装到当前用户目录，避免 Program Files 写权限问题，无需管理员权限
InstallDir "$LOCALAPPDATA\PhiloFTP"
InstallDirRegKey HKCU "${REG_UNINSTALL}" "InstallLocation"
VIProductVersion "__VERSION4__"
VIAddVersionKey "ProductName" "PhiloFTP 内网 FTP 服务器"
VIAddVersionKey "ProductVersion" "${APP_VERSION}"
VIAddVersionKey "FileDescription" "PhiloFTP 内网 FTP 服务器安装程序"
VIAddVersionKey "CompanyName" "${APP_PUBLISHER}"
VIAddVersionKey "LegalCopyright" "© ${APP_PUBLISHER}"
VIAddVersionKey "FileVersion" "${APP_VERSION}"

; ---------- Modern UI ----------
!define MUI_ABORTWARNING
!define MUI_ICON "philoftp.ico"
!define MUI_UNICON "philoftp.ico"
!define MUI_WELCOMEPAGE_TITLE "欢迎安装 PhiloFTP 内网 FTP 服务器"
!define MUI_WELCOMEPAGE_TEXT "本向导将引导您完成 PhiloFTP 的安装。$\r$\n$\r$\nPhiloFTP 是一个轻量的内网 FTP 服务器，内置 Web 管理端。$\r$\n$\r$\n安装完成后可通过 Web 管理页（http://本机:9090）进行配置。"

; 页面
!insertmacro MUI_PAGE_WELCOME
!insertmacro MUI_PAGE_LICENSE "LICENSE.txt"
!insertmacro MUI_PAGE_DIRECTORY
!insertmacro MUI_PAGE_INSTFILES
!define MUI_FINISHPAGE_RUN "$INSTDIR\philoftp.exe"
!define MUI_FINISHPAGE_RUN_TEXT "立即启动 PhiloFTP"
!define MUI_FINISHPAGE_SHOWREADME "$INSTDIR\README.txt"
!insertmacro MUI_PAGE_FINISH

!insertmacro MUI_UNPAGE_CONFIRM
!insertmacro MUI_UNPAGE_INSTFILES

!insertmacro MUI_LANGUAGE "SimpChinese"
!insertmacro MUI_LANGUAGE "English"

; ---------- 安装段 ----------
Section "PhiloFTP" SecMain
  SetOutPath "$INSTDIR"
  ; 主程序（重命名为 philoftp.exe）
  File "/oname=philoftp.exe" "__EXE_NAME__"
  ; 图标（供快捷方式用）
  File "philoftp.ico"
  ; 说明文档
  File "README.txt"

  ; 写入卸载信息到注册表（当前用户）
  WriteRegStr HKCU "${REG_UNINSTALL}" "DisplayName" "${APP_NAME}"
  WriteRegStr HKCU "${REG_UNINSTALL}" "DisplayVersion" "${APP_VERSION}"
  WriteRegStr HKCU "${REG_UNINSTALL}" "DisplayIcon" "$INSTDIR\philoftp.ico"
  WriteRegStr HKCU "${REG_UNINSTALL}" "Publisher" "${APP_PUBLISHER}"
  WriteRegStr HKCU "${REG_UNINSTALL}" "InstallLocation" "$INSTDIR"
  WriteRegStr HKCU "${REG_UNINSTALL}" "UninstallString" "$INSTDIR\Uninstall.exe"
  WriteRegStr HKCU "${REG_UNINSTALL}" "QuietUninstallString" "$INSTDIR\Uninstall.exe /S"
  WriteRegStr HKCU "${REG_UNINSTALL}" "URLInfoAbout" "${APP_WEB}"
  WriteRegDWORD HKCU "${REG_UNINSTALL}" "NoModify" 1
  WriteRegDWORD HKCU "${REG_UNINSTALL}" "NoRepair" 1
  WriteRegDWORD HKCU "${REG_UNINSTALL}" "EstimatedSize" 51200
  WriteUninstaller "$INSTDIR\Uninstall.exe"

  ; 桌面快捷方式（可选，MUI 无独立页，直接创建）
  CreateShortCut "$DESKTOP\PhiloFTP.lnk" "$INSTDIR\philoftp.exe" "" "$INSTDIR\philoftp.ico"
  ; 开始菜单快捷方式
  CreateDirectory "$SMPROGRAMS\PhiloFTP"
  CreateShortCut "$SMPROGRAMS\PhiloFTP\PhiloFTP.lnk" "$INSTDIR\philoftp.exe" "" "$INSTDIR\philoftp.ico"
  CreateShortCut "$SMPROGRAMS\PhiloFTP\卸载 PhiloFTP.lnk" "$INSTDIR\Uninstall.exe"

  ; 防火墙放行（需管理员，普通用户静默失败不影响安装）
  ExecWait 'netsh advfirewall firewall add rule name="PhiloFTP" dir=in action=allow program="$INSTDIR\philoftp.exe" enable=yes'
SectionEnd

; ---------- 卸载段 ----------
Section "Uninstall"
  ; 停止正在运行的进程
  ExecWait 'taskkill /f /im philoftp.exe'
  Sleep 500

  ; 删除快捷方式
  Delete "$DESKTOP\PhiloFTP.lnk"
  Delete "$SMPROGRAMS\PhiloFTP\PhiloFTP.lnk"
  Delete "$SMPROGRAMS\PhiloFTP\卸载 PhiloFTP.lnk"
  RMDir "$SMPROGRAMS\PhiloFTP"

  ; 删除防火墙规则
  ExecWait 'netsh advfirewall firewall delete rule name="PhiloFTP"'

  ; 删除注册表卸载信息
  DeleteRegKey HKCU "${REG_UNINSTALL}"
  DeleteRegKey HKCU "Software\PhiloFTP"

  ; 删除文件
  Delete "$INSTDIR\philoftp.exe"
  Delete "$INSTDIR\philoftp.ico"
  Delete "$INSTDIR\Uninstall.exe"
  Delete "$INSTDIR\README.txt"
  RMDir /r "$INSTDIR\data"
  RMDir /r "$INSTDIR\configs"
  RMDir "$INSTDIR"

SectionEnd
