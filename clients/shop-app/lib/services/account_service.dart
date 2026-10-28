import 'package:dio/dio.dart';
import 'package:quickeats_shop/core/constants/api_constants.dart';
import 'package:quickeats_shop/core/network/api_client.dart';
import 'package:quickeats_shop/models/organisation.dart';
import 'package:quickeats_shop/models/profile.dart';
import 'package:quickeats_shop/models/shop.dart';

/// Organisation-code lookup, the merchant's account, and their shop's details.
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

  /// Replaces the organisations the shop serves; returns the updated profile.
  Future<Profile> setShopOrganisations(List<String> codes) async {
    final resp = await _api.dio.put(ApiConstants.meOrganisations, data: {'organisationSlugs': codes});
    return Profile.fromJson(resp.data['data']);
  }

  Future<Shop> getShop(String shopId) async {
    final resp = await _api.dio.get(ApiConstants.shop(shopId));
    return Shop.fromJson(resp.data['data']);
  }

  Future<Shop> updateShop(String shopId, {String? name, String? description, bool? isOpen}) async {
    final resp = await _api.dio.patch(ApiConstants.shop(shopId), data: {
      if (name != null) 'name': name,
      if (description != null) 'description': description,
      if (isOpen != null) 'isOpen': isOpen,
    });
    return Shop.fromJson(resp.data['data']);
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
