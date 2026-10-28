import 'package:flutter/foundation.dart';
import 'package:quickeats_customer/models/profile.dart';
import 'package:quickeats_customer/services/account_service.dart';
import 'package:quickeats_customer/services/order_notification_service.dart';

class SettingsViewModel extends ChangeNotifier {
  final AccountService _account;
  final OrderNotificationService _notifications;

  SettingsViewModel({required AccountService accountService, required OrderNotificationService notifications})
      : _account = accountService, _notifications = notifications;

  Profile? _profile;
  bool _isLoading = false;
  String? _error;

  Profile? get profile => _profile;
  bool get isLoading => _isLoading;
  String? get error => _error;
  bool get orderAlertsEnabled => _notifications.alertsEnabled;

  Future<void> load() async {
    _isLoading = true; _error = null; notifyListeners();
    try {
      _profile = await _account.getProfile();
    } catch (e) {
      _error = apiErrorMessage(e, 'Could not load your account');
    }
    _isLoading = false; notifyListeners();
  }

  /// Returns null on success, or an error message to show.
  Future<String?> updateDisplayName(String name) async {
    try {
      _profile = await _account.updateDisplayName(name.trim());
      notifyListeners();
      return null;
    } catch (e) {
      return apiErrorMessage(e, 'Could not update your name');
    }
  }

  /// Returns null on success, or an error message to show.
  Future<String?> changePassword(String current, String next) async {
    try {
      await _account.changePassword(current, next);
      return null;
    } catch (e) {
      return apiErrorMessage(e, 'Could not change your password');
    }
  }

  Future<void> setOrderAlerts(bool enabled) async {
    await _notifications.setAlertsEnabled(enabled);
    notifyListeners();
  }

  void clear() { _profile = null; _error = null; }
}
