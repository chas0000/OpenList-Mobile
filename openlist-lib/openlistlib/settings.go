package openlistlib

import 'dart:io';

import 'package:flutter/material.dart';
import 'package:get/get.dart';
import 'package:file_picker/file_picker.dart';
import 'package:permission_handler/permission_handler.dart';
import 'package:path_provider/path_provider.dart';

import 'package:openlist_mobile/contant/native_bridge.dart';
import 'package:openlist_mobile/generated_api.dart';
import 'package:openlist_mobile/pages/settings/preference_widgets.dart';
import 'package:openlist_mobile/pages/settings/troubleshooting_page.dart';
import 'package:openlist_mobile/utils/language_controller.dart';
import '../../generated/l10n.dart';

class SettingsScreen extends StatefulWidget {
  const SettingsScreen({Key? key}) : super(key: key);

  @override
  State<SettingsScreen> createState() {
    return _SettingsScreenState();
  }
}

class _SettingsScreenState extends State<SettingsScreen> {
  late AppLifecycleListener _lifecycleListener;

  @override
  void initState() {
    _lifecycleListener = AppLifecycleListener(
      onResume: () async {
        final controller = Get.put(_SettingsController());
        controller.updateData();
      },
    );
    super.initState();
  }

  @override
  void dispose() {
    _lifecycleListener.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final controller = Get.put(_SettingsController());

    return Scaffold(
      body: Obx(
        () => ListView(
          children: [
            // 权限部分
            Visibility(
              visible: !controller._managerStorageGranted.value ||
                  !controller._notificationGranted.value ||
                  !controller._storageGranted.value,
              child: DividerPreference(title: S.of(context).importantSettings),
            ),
            Visibility(
              visible: !controller._managerStorageGranted.value,
              child: BasicPreference(
                title: S.of(context).grantManagerStoragePermission,
                subtitle: S.of(context).grantStoragePermissionDesc,
                onTap: () {
                  Permission.manageExternalStorage.request();
                },
              ),
            ),
            Visibility(
              visible: !controller._storageGranted.value,
              child: BasicPreference(
                title: S.of(context).grantStoragePermission,
                subtitle: S.of(context).grantStoragePermissionDesc,
                onTap: () {
                  Permission.storage.request();
                },
              ),
            ),
            Visibility(
              visible: !controller._notificationGranted.value,
              child: BasicPreference(
                title: S.of(context).grantNotificationPermission,
                subtitle: S.of(context).grantNotificationPermissionDesc,
                onTap: () {
                  Permission.notification.request();
                },
              ),
            ),

            DividerPreference(title: S.of(context).general),

            // 语言
            BasicPreference(
              title: S.of(context).language,
              subtitle: _getLanguageDisplayName(),
              leading: const Icon(Icons.language),
              onTap: () {
                _showLanguageSelectionDialog(context);
              },
            ),

            // 自动更新
            SwitchPreference(
              title: S.of(context).autoCheckForUpdates,
              subtitle: S.of(context).autoCheckForUpdatesDesc,
              icon: const Icon(Icons.system_update),
              value: controller.autoUpdate,
              onChanged: (value) {
                controller.autoUpdate = value;
              },
            ),
            // 保持唤醒
            SwitchPreference(
              title: S.of(context).wakeLock,
              subtitle: S.of(context).wakeLockDesc,
              icon: const Icon(Icons.screen_lock_portrait),
              value: controller.wakeLock,
              onChanged: (value) {
                controller.wakeLock = value;
              },
            ),
            // 开机启动
            SwitchPreference(
              title: S.of(context).bootAutoStartService,
              subtitle: S.of(context).bootAutoStartServiceDesc,
              icon: const Icon(Icons.power_settings_new),
              value: controller.startAtBoot,
              onChanged: (value) {
                controller.startAtBoot = value;
              },
            ),
            // 自动打开网页
            SwitchPreference(
              title: S.of(context).autoStartWebPage,
              subtitle: S.of(context).autoStartWebPageDesc,
              icon: const Icon(Icons.open_in_browser),
              value: controller.autoStartWebPage,
              onChanged: (value) {
                controller.autoStartWebPage = value;
              },
            ),

            // ✅ 跨平台数据目录选择
            BasicPreference(
              title: S.of(context).dataDirectory,
              subtitle: controller.dataDir.isEmpty
                  ? S.of(context).setDefaultDirectory
                  : controller.dataDir,
              leading: const Icon(Icons.folder),
              onTap: () async {
                String? selectedPath;

                if (Platform.isIOS) {
                  // iOS 使用沙盒目录
                  final dir = await getApplicationDocumentsDirectory();
                  selectedPath = dir.path;

                  Get.snackbar(
                    S.of(context).dataDirectory,
                    'iOS 使用默认应用目录：\n${dir.path}',
                    duration: const Duration(seconds: 4),
                  );
                } else if (Platform.isAndroid ||
                    Platform.isWindows ||
                    Platform.isMacOS ||
                    Platform.isLinux) {
                  // 桌面和 Android 使用目录选择器
                  selectedPath = await FilePicker.platform.getDirectoryPath();

                  if (selectedPath == null) {
                    Get.showSnackbar(GetSnackBar(
                      message: S.current.setDefaultDirectory,
                      duration: const Duration(seconds: 3),
                      mainButton: TextButton(
                        onPressed: () {
                          controller.setDataDir("");
                          Get.back();
                        },
                        child: Text(S.current.confirm),
                      ),
                    ));
                    return;
                  }
                } else {
                  // Web 或未知平台
                  Get.snackbar(
                    S.of(context).dataDirectory,
                    '当前平台不支持目录选择，将使用默认目录',
                  );
                  selectedPath = "";
                }

                await controller.setDataDir(selectedPath);
              },
            ),

            DividerPreference(title: S.of(context).uiSettings),
            SwitchPreference(
              icon: const Icon(Icons.pan_tool_alt_outlined),
              title: S.of(context).silentJumpApp,
              subtitle: S.of(context).silentJumpAppDesc,
              value: controller.silentJumpApp,
              onChanged: (value) {
                controller.silentJumpApp = value;
              },
            ),

            // 故障排查
            BasicPreference(
              title: S.of(context).troubleshooting,
              subtitle: S.of(context).troubleshootingDesc,
              leading: const Icon(Icons.help_outline),
              onTap: () {
                Navigator.push(
                  context,
                  MaterialPageRoute(
                    builder: (context) => const TroubleshootingPage(),
                  ),
                );
              },
            ),
          ],
        ),
      ),
    );
  }

  String _getLanguageDisplayName() {
    final languageController = Get.find<LanguageController>();
    final currentOption = languageController.currentLanguageOption;

    switch (currentOption.name) {
      case 'followSystem':
        return S.of(context).followSystem;
      case 'simplifiedChinese':
        return S.of(context).simplifiedChinese;
      case 'english':
        return S.of(context).english;
      default:
        return currentOption.name;
    }
  }

  void _showLanguageSelectionDialog(BuildContext context) {
    showDialog(
      context: context,
      builder: (BuildContext context) {
        return AlertDialog(
          title: Text(S.of(context).languageSettings),
          content: SingleChildScrollView(
            child: LanguageSelector(
              onLanguageChanged: () {
                Navigator.of(context).pop();
                setState(() {}); // 刷新语言
              },
            ),
          ),
          actions: [
            TextButton(
              onPressed: () {
                Navigator.of(context).pop();
              },
              child: Text(S.of(context).cancel),
            ),
          ],
        );
      },
    );
  }
}

class _SettingsController extends GetxController {
  final _dataDir = "".obs;
  final _autoUpdate = true.obs;
  final _managerStorageGranted = true.obs;
  final _notificationGranted = true.obs;
  final _storageGranted = true.obs;

  setDataDir(String value) async {
    NativeBridge.appConfig.setDataDir(value);
    _dataDir.value = await NativeBridge.appConfig.getDataDir();
  }

  String get dataDir => _dataDir.value;

  set autoUpdate(value) {
    _autoUpdate.value = value;
    NativeBridge.appConfig.setAutoCheckUpdateEnabled(value);
  }

  bool get autoUpdate => _autoUpdate.value;

  final _wakeLock = true.obs;
  set wakeLock(value) {
    _wakeLock.value = value;
    NativeBridge.appConfig.setWakeLockEnabled(value);
  }

  bool get wakeLock => _wakeLock.value;

  final _autoStart = true.obs;
  set startAtBoot(value) {
    _autoStart.value = value;
    NativeBridge.appConfig.setStartAtBootEnabled(value);
  }

  bool get startAtBoot => _autoStart.value;

  final _autoStartWebPage = false.obs;
  set autoStartWebPage(value) {
    _autoStartWebPage.value = value;
    NativeBridge.appConfig.setAutoOpenWebPageEnabled(value);
  }

  bool get autoStartWebPage => _autoStartWebPage.value;

  final _silentJumpApp = false.obs;
  bool get silentJumpApp => _silentJumpApp.value;
  set silentJumpApp(value) {
    _silentJumpApp.value = value;
    NativeBridge.appConfig.setSilentJumpAppEnabled(value);
  }

  @override
  void onInit() async {
    updateData();
    super.onInit();
  }

  void updateData() async {
    final cfg = AppConfig();
    cfg.isAutoCheckUpdateEnabled().then((value) => autoUpdate = value);
    cfg.isWakeLockEnabled().then((value) => wakeLock = value);
    cfg.isStartAtBootEnabled().then((value) => startAtBoot = value);
    cfg.isAutoOpenWebPageEnabled().then((value) => autoStartWebPage = value);
    cfg.isSilentJumpAppEnabled().then((value) => silentJumpApp = value);

    _dataDir.value = await cfg.getDataDir();

    final sdk = await NativeBridge.common.getDeviceSdkInt();
    // Android 11+
    if (sdk >= 30) {
      _managerStorageGranted.value =
          await Permission.manageExternalStorage.isGranted;
    } else {
      _managerStorageGranted.value = true;
      _storageGranted.value = await Permission.storage.isGranted;
    }

    // Android 12+
    if (sdk >= 32) {
      _notificationGranted.value = await Permission.notification.isGranted;
    } else {
      _notificationGranted.value = true;
    }
  }
}
