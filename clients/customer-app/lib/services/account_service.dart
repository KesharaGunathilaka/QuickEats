import 'package:dio/dio.dart';
import 'package:quickeats_customer/core/constants/api_constants.dart';
import 'package:quickeats_customer/core/network/api_client.dart';
import 'package:quickeats_customer/models/organisation.dart';
import 'package:quickeats_customer/models/profile.dart';

/// Organisation-code lookup (sign-up) and the signed-in user's own account (settings).
class AccountService {
  final ApiClient _api;
  AccountService({required ApiClient apiClient}) : _api = apiClient;

  /// Returns the organisation for a code, or null if no such code exists.
  Future<Organisation?> lookupOrganisation(String code) async {
    try {
      final resp = await _api.dio.get(ApiConstants.organisation(code.trim().toLowerCase()));
      return Organisation.fromJson(resp.data['data']);
    } on DioException catch (e) {
      if (e.response?.statusCode == 404) return null;
      rethrow;
    }
  }

  Future<Profile> getProfile() async {
    final resp = await _api.dio.get(ApiConstants.me);
    return Profile.fromJson(resp.data['data']);
  }

  Future<Profile> updateDisplayName(String name) async {
    final resp = await _api.dio.patch(ApiConstants.me, data: {'displayName': name});
    return Profile.fromJson(resp.data['data']);
  }

  Future<void> changePassword(String current, String next) async {
    await _api.dio.post(ApiConstants.mePassword, data: {'currentPassword': current, 'newPassword': next});
  }
}

/// Backend errors look like {"error": {"code": ..., "message": ...}}; returns a readable message.
String apiErrorMessage(Object e, String fallback) {
  if (e is DioException) {
    final data = e.response?.data;
    if (data is Map && data['error'] is Map && data['error']['message'] is String) {
      final msg = data['error']['message'] as String;
      return msg.isEmpty ? fallback : msg[0].toUpperCase() + msg.substring(1);
    }
    if (e.type == DioExceptionType.connectionTimeout || e.type == DioExceptionType.connectionError) {
      return 'Can\'t reach the server. Check your connection.';
    }
  }
  return fallback;
}
