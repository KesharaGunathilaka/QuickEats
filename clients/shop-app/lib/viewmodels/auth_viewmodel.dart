import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';
import 'package:quickeats_shop/core/di/service_locator.dart';
import 'package:quickeats_shop/core/network/ws_client.dart';
import 'package:quickeats_shop/services/auth_service.dart';

class AuthViewModel extends ChangeNotifier {
  final AuthService _authService;

  AuthViewModel({required AuthService authService}) : _authService = authService;

  bool _isLoading = false;
  String? _error;

  bool get isLoading => _isLoading;
  String? get error => _error;
  bool get isAuthenticated => _authService.isAuthenticated;
  String? get shopId => _authService.shopId;

  Future<bool> login(String email, String password) async {
    _isLoading = true;
    _error = null;
    notifyListeners();

    try {
      await _authService.login(email, password);
      _isLoading = false;
      notifyListeners();
      return true;
    } catch (e) {
      _isLoading = false;
      _error = 'Invalid email or password';
      notifyListeners();
      return false;
    }
  }

  Future<bool> registerShop({
    required List<String> orgSlugs, required String email, required String password,
    required String displayName, required String shopName, String shopDescription = '',
  }) async {
    _isLoading = true;
    _error = null;
    notifyListeners();

    try {
      await _authService.registerShop(
        orgSlugs: orgSlugs, email: email, password: password,
        displayName: displayName, shopName: shopName, shopDescription: shopDescription,
      );
      _isLoading = false;
      notifyListeners();
      return true;
    } catch (e) {
      _isLoading = false;
      _error = _serverMessage(e) ?? 'Registration failed. Check your details.';
      notifyListeners();
      return false;
    }
  }

  Future<void> logout() async {
    getIt<WSClient>().disconnect(); // stop live orders for the old session
    await _authService.logout();
    _error = null;
    notifyListeners();
  }

  void clearError() {
    _error = null;
    notifyListeners();
  }

  // Backend errors look like {"error": {"code": ..., "message": ...}}
  String? _serverMessage(Object e) {
    if (e is! DioException) return null;
    final data = e.response?.data;
    if (data is Map && data['error'] is Map) return data['error']['message'] as String?;
    return null;
  }
}
