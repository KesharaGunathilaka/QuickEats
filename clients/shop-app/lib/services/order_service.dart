import 'package:quickeats_shop/core/constants/api_constants.dart';
import 'package:quickeats_shop/core/network/api_client.dart';
import 'package:quickeats_shop/models/order.dart';

class OrderService {
  final ApiClient _apiClient;

  OrderService({required ApiClient apiClient}) : _apiClient = apiClient;

  Future<List<Order>> getShopOrders(String shopId) async {
    final response = await _apiClient.dio.get(ApiConstants.shopOrders(shopId));
    final list = response.data['data'] as List<dynamic>? ?? [];
    return list.map((e) => Order.fromJson(e as Map<String, dynamic>)).toList();
  }

  Future<Order> getOrder(String orderId) async {
    final response = await _apiClient.dio.get(ApiConstants.orderById(orderId));
    return Order.fromJson(response.data['data']);
  }

  Future<void> updateStatus(String orderId, String status) async {
    await _apiClient.dio.patch(ApiConstants.updateOrderStatus(orderId), data: {
      'status': status,
    });
  }

  Future<void> rejectOrder(String orderId, String reason) async {
    await _apiClient.dio.post(ApiConstants.rejectOrder(orderId), data: {
      'reason': reason,
    });
  }
}
