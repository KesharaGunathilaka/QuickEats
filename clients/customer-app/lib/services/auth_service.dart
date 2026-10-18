import 'dart:convert';
import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'package:quickeats_customer/core/constants/api_constants.dart';
import 'package:quickeats_customer/models/auth_tokens.dart';
import 'package:quickeats_customer/models/user.dart';

class AuthService extends ChangeNotifier {
  final _storage = const FlutterSecureStorage();
  final _rawDio = Dio(BaseOptions(baseUrl: ApiConstants.baseUrl));

  String? _accessToken;
  String? _refreshTokenStr;
  User? _currentUser;

  String? get accessToken => _accessToken;
  String? get refreshToken => _refreshTokenStr;
  User? get currentUser => _currentUser;
  bool get isAuthenticated => _accessToken != null;

  Future<void> init() async {
    _accessToken = await _storage.read(key: 'access_token');
    _refreshTokenStr = await _storage.read(key: 'refresh_token');
    final uj = await _storage.read(key: 'user');
    if (uj != null) _currentUser = User.fromJson(jsonDecode(uj));
    notifyListeners();
  }

  Future<User> login(String email, String password) async {
    final resp = await _rawDio.post(ApiConstants.login, data: {'email': email, 'password': password});
    final tokens = AuthTokens.fromJson(resp.data['data']);
    await _saveTokens(tokens);
    _currentUser = _decodeUser(tokens.accessToken, email);
    await _storage.write(key: 'user', value: jsonEncode(_currentUser!.toJson()));
    notifyListeners();
    return _currentUser!;
  }

  Future<void> register(String orgSlug, String email, String password, String displayName, {String? inviteCode}) async {
    await _rawDio.post(ApiConstants.register, data: {
      'organisationSlug': orgSlug, 'email': email, 'password': password, 'displayName': displayName,
      if (inviteCode != null) 'inviteCode': inviteCode,
    });
  }

  Future<bool> refreshTokens() async {
    if (_refreshTokenStr == null) return false;
    try {
      final resp = await _rawDio.post(ApiConstants.refresh, data: {'refreshToken': _refreshTokenStr});
      await _saveTokens(AuthTokens.fromJson(resp.data['data']));
      notifyListeners();
      return true;
    } catch (_) { await logout(); return false; }
  }

  Future<void> logout() async {
    _accessToken = null; _refreshTokenStr = null; _currentUser = null;
    await _storage.deleteAll();
    notifyListeners();
  }

  Future<void> _saveTokens(AuthTokens t) async {
    _accessToken = t.accessToken; _refreshTokenStr = t.refreshToken;
    await _storage.write(key: 'access_token', value: t.accessToken);
    await _storage.write(key: 'refresh_token', value: t.refreshToken);
  }

  User _decodeUser(String token, String email) {
    try {
      final p = jsonDecode(utf8.decode(base64Url.decode(base64Url.normalize(token.split('.')[1]))));
      return User(id: p['sub'] ?? '', organisationId: p['org'] ?? '', email: email, displayName: email.split('@').first, role: p['role'] ?? '', shopId: p['shop']);
    } catch (_) { return User(id: '', organisationId: '', email: email, displayName: email.split('@').first, role: 'customer'); }
  }
}
