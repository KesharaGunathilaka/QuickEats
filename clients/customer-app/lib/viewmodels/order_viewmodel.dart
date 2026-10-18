import 'dart:async';
import 'package:flutter/foundation.dart';
import 'package:quickeats_customer/models/order.dart';
import 'package:quickeats_customer/models/cart_item.dart';
import 'package:quickeats_customer/services/order_service.dart';
import 'package:uuid/uuid.dart';

class OrderViewModel extends ChangeNotifier {
  final OrderService _orderService;
  OrderViewModel({required OrderService orderService}) : _orderService = orderService;

  List<Order> _orders = [];
  Order? _activeOrder;
  bool _isLoading = false;
  bool _isPlacing = false;
  String? _error;
  Timer? _pollTimer;

  List<Order> get orders => _orders;
  Order? get activeOrder => _activeOrder;
  bool get isLoading => _isLoading;
  bool get isPlacing => _isPlacing;
  String? get error => _error;

  Future<Order?> placeOrder({required String shopId, required String paymentMethod, required List<CartItem> items}) async {
    _isPlacing = true; _error = null; notifyListeners();
    try {
      final order = await _orderService.placeOrder(
        shopId: shopId, idempotencyKey: const Uuid().v4(), paymentMethod: paymentMethod, items: items,
      );
      _activeOrder = order;
      _startPolling(order.id);
      _isPlacing = false; notifyListeners();
      return order;
    } catch (e) {
      _isPlacing = false; _error = 'Failed to place order'; notifyListeners();
      return null;
    }
  }

  Future<void> loadOrders() async {
    _isLoading = true; notifyListeners();
    try { _orders = await _orderService.getMyOrders(); _orders.sort((a, b) => b.createdAt.compareTo(a.createdAt));
    } catch (e) { _error = 'Failed to load orders'; }
    _isLoading = false; notifyListeners();
  }

  Future<void> trackOrder(String orderId) async {
    try { _activeOrder = await _orderService.getOrder(orderId); _startPolling(orderId); notifyListeners();
    } catch (e) { _error = 'Failed to load order'; notifyListeners(); }
  }

  Future<bool> cancelOrder(String orderId) async {
    try { await _orderService.cancelOrder(orderId); await trackOrder(orderId); return true;
    } catch (e) { _error = 'Failed to cancel'; notifyListeners(); return false; }
  }

  void _startPolling(String orderId) {
    _pollTimer?.cancel();
    _pollTimer = Timer.periodic(const Duration(seconds: 5), (_) async {
      try {
        _activeOrder = await _orderService.getOrder(orderId);
        notifyListeners();
        if (_activeOrder!.status.isTerminal) _pollTimer?.cancel();
      } catch (_) {}
    });
  }

  void stopPolling() { _pollTimer?.cancel(); }

  @override
  void dispose() { _pollTimer?.cancel(); super.dispose(); }
}
