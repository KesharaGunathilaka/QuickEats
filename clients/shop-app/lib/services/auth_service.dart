import 'dart:convert';
import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'package:quickeats_shop/core/constants/api_constants.dart';
import 'package:quickeats_shop/models/auth_tokens.dart';
import 'package:quickeats_shop/models/user.dart';

class AuthService extends ChangeNotifier {
  final FlutterSecureStorage _storage = const FlutterSecureStorage();
  final Dio _rawDio = Dio(BaseOptions(baseUrl: ApiConstants.baseUrl));

  String? _accessToken;
  String? _refreshTokenStr;
  User? _currentUser;

  String? get accessToken => _accessToken;
  String? get refreshToken => _refreshTokenStr;
  User? get currentUser => _currentUser;
  bool get isAuthenticated => _accessToken != null;
  String? get shopId => _currentUser?.shopId ?? _parseShopFromToken();

  Future<void> init() async {
    _accessToken = await _storage.read(key: 'access_token');
    _refreshTokenStr = await _storage.read(key: 'refresh_token');
    final userJson = await _storage.read(key: 'user');
    if (userJson != null) {
      _currentUser = User.fromJson(jsonDecode(userJson));
    }
    notifyListeners();
  }

  Future<User> login(String email, String password) async {
    final response = await _rawDio.post(ApiConstants.login, data: {
      'email': email,
      'password': password,
    });

    final data = response.data['data'];
    final tokens = AuthTokens.fromJson(data);
    await _saveTokens(tokens);

    // Decode user from JWT payload
    _currentUser = _decodeUserFromToken(tokens.accessToken);
    await _storage.write(key: 'user', value: jsonEncode({
      'id': _currentUser!.id,
      'organisationId': _currentUser!.organisationId,
      'email': email,
      'displayName': email.split('@').first,
      'role': _currentUser!.role,
      'shopId': _currentUser!.shopId,
    }));

    notifyListeners();
    return _currentUser!;
  }

  /// Creates a shop-owner account together with its shop. Sign in afterwards.
  Future<void> registerShop({
    required String orgSlug, required String email, required String password,
    required String displayName, required String shopName, String shopDescription = '',
  }) async {
    await _rawDio.post(ApiConstants.registerShop, data: {
      'organisationSlug': orgSlug,
      'email': email,
      'password': password,
      'displayName': displayName,
      'shopName': shopName,
      'shopDescription': shopDescription,
    });
  }

  Future<bool> refreshTokens() async {
    if (_refreshTokenStr == null) return false;
    try {
      final response = await _rawDio.post(ApiConstants.refresh, data: {
        'refreshToken': _refreshTokenStr,
      });
      final data = response.data['data'];
      final tokens = AuthTokens.fromJson(data);
      await _saveTokens(tokens);
      notifyListeners();
      return true;
    } catch (e) {
      await logout();
      return false;
    }
  }

  Future<void> logout() async {
    _accessToken = null;
    _refreshTokenStr = null;
    _currentUser = null;
    await _storage.deleteAll();
    notifyListeners();
  }

  Future<void> _saveTokens(AuthTokens tokens) async {
    _accessToken = tokens.accessToken;
    _refreshTokenStr = tokens.refreshToken;
    await _storage.write(key: 'access_token', value: tokens.accessToken);
    await _storage.write(key: 'refresh_token', value: tokens.refreshToken);
  }

  User? _decodeUserFromToken(String token) {
    try {
      final parts = token.split('.');
      if (parts.length != 3) return null;
      final payload = utf8.decode(base64Url.decode(base64Url.normalize(parts[1])));
      final map = jsonDecode(payload) as Map<String, dynamic>;
      return User(
        id: map['sub'] ?? '',
        organisationId: map['org'] ?? '',
        email: '',
        displayName: '',
        role: map['role'] ?? '',
        shopId: map['shop'],
      );
    } catch (_) {
      return null;
    }
  }

  String? _parseShopFromToken() {
    if (_accessToken == null) return null;
    final user = _decodeUserFromToken(_accessToken!);
    return user?.shopId;
  }
}
