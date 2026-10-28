import 'package:flutter/foundation.dart';
import 'package:quickeats_shop/models/organisation.dart';
import 'package:quickeats_shop/models/profile.dart';
import 'package:quickeats_shop/models/shop.dart';
import 'package:quickeats_shop/services/account_service.dart';

/// Merchant settings: account, shop details, open/closed, and organisation codes.
/// Mutating methods return null on success or a message to show.
class SettingsViewModel extends ChangeNotifier {
  final AccountService _account;
  SettingsViewModel({required AccountService accountService}) : _account = accountService;

  Profile? _profile;
  Shop? _shop;
  bool _isLoading = false;
  String? _error;

  Profile? get profile => _profile;
  Shop? get shop => _shop;
  bool get isLoading => _isLoading;
  String? get error => _error;

  Future<void> load() async {
    _isLoading = true; _error = null; notifyListeners();
    try {
      _profile = await _account.getProfile();
      final shopId = _profile!.shopId;
      _shop = shopId == null ? null : await _account.getShop(shopId);
    } catch (e) {
      _error = apiErrorMessage(e, 'Could not load your account');
    }
    _isLoading = false; notifyListeners();
  }

  Future<String?> updateDisplayName(String name) => _run(() async {
    _profile = await _account.updateDisplayName(name);
  }, 'Could not update your name');

  Future<String?> changePassword(String current, String next) =>
      _run(() => _account.changePassword(current, next), 'Could not change your password');

  Future<String?> updateShopDetails(String name, String description) => _run(() async {
    _shop = await _account.updateShop(_shop!.id, name: name, description: description);
  }, 'Could not update the shop');

  Future<String?> setOpen(bool isOpen) => _run(() async {
    _shop = await _account.updateShop(_shop!.id, isOpen: isOpen);
  }, 'Could not update the shop');

  Future<String?> saveOrganisations(List<Organisation> orgs) => _run(() async {
    _profile = await _account.setShopOrganisations(orgs.map((o) => o.slug).toList());
  }, 'Could not update organisation codes');

  Future<String?> _run(Future<void> Function() action, String fallback) async {
    try {
      await action();
      notifyListeners();
      return null;
    } catch (e) {
      return apiErrorMessage(e, fallback);
    }
  }

  void clear() { _profile = null; _shop = null; _error = null; }
}
