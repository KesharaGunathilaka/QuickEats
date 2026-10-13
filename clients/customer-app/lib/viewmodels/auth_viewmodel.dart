import 'package:flutter/foundation.dart';
import 'package:quickeats_customer/services/auth_service.dart';

class AuthViewModel extends ChangeNotifier {
  final AuthService _auth;
  AuthViewModel({required AuthService authService}) : _auth = authService;

  bool _isLoading = false;
  String? _error;
  bool get isLoading => _isLoading;
  String? get error => _error;
  bool get isAuthenticated => _auth.isAuthenticated;

  Future<bool> login(String email, String password) async {
    _isLoading = true; _error = null; notifyListeners();
    try { await _auth.login(email, password); _isLoading = false; notifyListeners(); return true;
    } catch (e) { _isLoading = false; _error = 'Invalid email or password'; notifyListeners(); return false; }
  }

  Future<bool> register(String orgSlug, String email, String password, String displayName) async {
    _isLoading = true; _error = null; notifyListeners();
    try { await _auth.register(orgSlug, email, password, displayName); _isLoading = false; notifyListeners(); return true;
    } catch (e) { _isLoading = false; _error = 'Registration failed. Check your details.'; notifyListeners(); return false; }
  }

  Future<void> logout() async { await _auth.logout(); notifyListeners(); }
}
