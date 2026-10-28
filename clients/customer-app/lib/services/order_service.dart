import 'package:quickeats_customer/core/constants/api_constants.dart';
import 'package:quickeats_customer/core/network/api_client.dart';
import 'package:quickeats_customer/models/order.dart';
import 'package:quickeats_customer/models/cart_item.dart';

class OrderService {
  final ApiClient _api;
  OrderService({required ApiClient apiClient}) : _api = apiClient;

  /// Places an order paid with the customer's saved card.
  Future<Order> placeOrder({required String shopId, required String idempotencyKey, required List<CartItem> items}) async {
    final resp = await _api.dio.post(ApiConstants.orders, data: {
      'shopId': shopId,
      'idempotencyKey': idempotencyKey,
      'paymentMethod': 'card',
      'items': items.map((c) => {
        'menuItemId': c.menuItem.id,
        'itemName': c.menuItem.name,
        'quantity': c.quantity,
        'unitPrice': c.menuItem.price,
      }).toList(),
    });
    return Order.fromJson(resp.data['data']);
  }

  Future<List<Order>> getMyOrders() async {
    final resp = await _api.dio.get(ApiConstants.orders);
    return (resp.data['data'] as List? ?? []).map((e) => Order.fromJson(e)).toList();
  }

  Future<Order> getOrder(String id) async {
    final resp = await _api.dio.get(ApiConstants.orderById(id));
    return Order.fromJson(resp.data['data']);
  }

  Future<void> cancelOrder(String id) async {
    await _api.dio.post(ApiConstants.cancelOrder(id), data: {'reason': 'customer_requested'});
  }
}
