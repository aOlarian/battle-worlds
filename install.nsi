Unicode true
Name "Battle Worlds"
OutFile "battle-worlds-install.exe"
InstallDir "$LOCALAPPDATA\Programs\Battle Worlds"
RequestExecutionLevel user

Page directory
Page instfiles
UninstPage uninstConfirm
UninstPage instfiles

Section
  SetShellVarContext current
  SetOutPath "$INSTDIR"
  File "battle-worlds.exe"
  File "README.md"
  File "LICENSE"
  File /oname=AUDIO-CREDITS.txt "assets\sfx\CREDITS.txt"
  File /oname=LICENSE-kenney.txt "assets\sfx\LICENSE-kenney.txt"
  SetOutPath "$INSTDIR\assets\screenshots"
  File "assets\screenshots\coast.png"
  File "assets\screenshots\plaza.png"
  File "assets\screenshots\monster-world.png"
  SetOutPath "$INSTDIR"
  WriteUninstaller "$INSTDIR\Uninstall.exe"
  CreateDirectory "$SMPROGRAMS\Battle Worlds"
  CreateShortcut "$SMPROGRAMS\Battle Worlds\Battle Worlds.lnk" "$INSTDIR\battle-worlds.exe"
  WriteRegStr HKCU "Software\Microsoft\Windows\CurrentVersion\Uninstall\BattleWorlds" "DisplayName" "Battle Worlds"
  WriteRegStr HKCU "Software\Microsoft\Windows\CurrentVersion\Uninstall\BattleWorlds" "UninstallString" '"$INSTDIR\Uninstall.exe"'
SectionEnd

Section "Uninstall"
  SetShellVarContext current
  Delete "$SMPROGRAMS\Battle Worlds\Battle Worlds.lnk"
  RMDir "$SMPROGRAMS\Battle Worlds"
  Delete "$INSTDIR\battle-worlds.exe"
  Delete "$INSTDIR\README.md"
  Delete "$INSTDIR\LICENSE"
  Delete "$INSTDIR\AUDIO-CREDITS.txt"
  Delete "$INSTDIR\LICENSE-kenney.txt"
  Delete "$INSTDIR\assets\screenshots\coast.png"
  Delete "$INSTDIR\assets\screenshots\plaza.png"
  Delete "$INSTDIR\assets\screenshots\monster-world.png"
  RMDir "$INSTDIR\assets\screenshots"
  RMDir "$INSTDIR\assets"
  Delete "$INSTDIR\Uninstall.exe"
  RMDir "$INSTDIR"
  DeleteRegKey HKCU "Software\Microsoft\Windows\CurrentVersion\Uninstall\BattleWorlds"
SectionEnd
