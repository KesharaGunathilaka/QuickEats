import 'package:quickeats_shop/core/constants/api_constants.dart';
import 'package:quickeats_shop/core/network/api_client.dart';
import 'package:quickeats_shop/models/menu_item.dart';

class MenuService {
  final ApiClient _apiClient;

  MenuService({required ApiClient apiClient}) : _apiClient = apiClient;

  Future<List<MenuItem>> getMenuItems(String shopId) async {
    final response = await _apiClient.dio.get(ApiConstants.shopMenu(shopId));
    final list = response.data['data'] as List<dynamic>? ?? [];
    return list.map((e) => MenuItem.fromJson(e as Map<String, dynamic>)).toList();
  }

  Future<MenuItem> createItem(String shopId, {required String name, required String description, required int price, required String category, required int stock}) async {
    final response = await _apiClient.dio.post(ApiConstants.createMenuItem(shopId), data: {
      'name': name, 'description': description, 'price': price, 'category': category, 'stock': stock,
    });
    return MenuItem.fromJson(response.data['data']);
  }

  Future<void> updateStock(String itemId, int stock) async {
    await _apiClient.dio.patch(ApiConstants.updateStock(itemId), data: {'stock': stock});
  }

  Future<void> deleteItem(String itemId) async {
    await _apiClient.dio.delete(ApiConstants.menuItem(itemId));
  }
}
