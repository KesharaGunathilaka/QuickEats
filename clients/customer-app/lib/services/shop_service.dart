import 'package:quickeats_customer/core/constants/api_constants.dart';
import 'package:quickeats_customer/core/network/api_client.dart';
import 'package:quickeats_customer/models/shop.dart';
import 'package:quickeats_customer/models/menu_item.dart';

class ShopService {
  final ApiClient _api;
  ShopService({required ApiClient apiClient}) : _api = apiClient;

  Future<List<Shop>> getShops() async {
    final resp = await _api.dio.get(ApiConstants.shops);
    return (resp.data['data'] as List? ?? []).map((e) => Shop.fromJson(e)).toList();
  }

  Future<Shop> getShop(String id) async {
    final resp = await _api.dio.get(ApiConstants.shopById(id));
    return Shop.fromJson(resp.data['data']);
  }

  Future<List<MenuItem>> getMenu(String shopId) async {
    final resp = await _api.dio.get(ApiConstants.shopMenu(shopId));
    return (resp.data['data'] as List? ?? []).map((e) => MenuItem.fromJson(e)).toList();
  }
}
