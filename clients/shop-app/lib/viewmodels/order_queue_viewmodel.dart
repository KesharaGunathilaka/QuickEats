import 'dart:async';
import 'package:flutter/foundation.dart';
import 'package:quickeats_shop/core/constants/order_status.dart';
import 'package:quickeats_shop/core/network/ws_client.dart';
import 'package:quickeats_shop/models/order.dart';
import 'package:quickeats_shop/services/account_service.dart';
import 'package:quickeats_shop/services/auth_service.dart';
import 'package:quickeats_shop/services/order_service.dart';

class OrderQueueViewModel extends ChangeNotifier {
  final OrderService _orderService;
  final AuthService _authService;
  final WSClient _wsClient;

  OrderQueueViewModel({
    required OrderService orderService,
    required AuthService authService,
    required WSClient wsClient,
  }) : _orderService = orderService, _authService = authService, _wsClient = wsClient;

  List<Order> _orders = [];
  final _notices = StreamController<String>.broadcast();
  bool _isLoading = false;
  String? _error;
  OrderStatus _filterStatus = OrderStatus.confirmed;

  List<Order> get orders => _orders.where((o) => o.status.queueTab == _filterStatus).toList();
  List<Order> get allOrders => _orders;
  bool get isLoading => _isLoading;
  String? get error => _error;
  OrderStatus get filterStatus => _filterStatus;
  String? get shopId => _authService.shopId;
  /// One-off messages for the screen to show (e.g. a customer cancelled an order).
  Stream<String> get notices => _notices.stream;

  int countFor(OrderStatus tab) => _orders.where((o) => o.status.queueTab == tab).length;

  Future<void> init() async {
    await loadOrders();
    _connectWebSocket();
  }

  void setFilter(OrderStatus status) {
    _filterStatus = status;
    notifyListeners();
  }

  Future<void> loadOrders() async {
    final sid = shopId;
    if (sid == null) return;

    _isLoading = true;
    _error = null;
    notifyListeners();

    try {
      _orders = await _orderService.getShopOrders(sid);
      // Sort: newest first
      _orders.sort((a, b) => b.createdAt.compareTo(a.createdAt));
      _isLoading = false;
      notifyListeners();
    } catch (e) {
      _isLoading = false;
      _error = 'Failed to load orders: $e';
      notifyListeners();
    }
  }

  Future<bool> updateStatus(String orderId, OrderStatus newStatus) async {
    try {
      await _orderService.updateStatus(orderId, newStatus.value);
      final idx = _orders.indexWhere((o) => o.id == orderId);
      if (idx >= 0) {
        _orders[idx].status = newStatus;
        notifyListeners();
      }
      return true;
    } catch (e) {
      final message = apiErrorMessage(e, 'Failed to update the order');
      await loadOrders(); // the customer may have cancelled it meanwhile
      _error = message;
      notifyListeners();
      return false;
    }
  }

  Future<bool> rejectOrder(String orderId, String reason) async {
    try {
      await _orderService.rejectOrder(orderId, reason);
      final idx = _orders.indexWhere((o) => o.id == orderId);
      if (idx >= 0) {
        _orders[idx].status = OrderStatus.rejected;
        notifyListeners();
      }
      return true;
    } catch (e) {
      final message = apiErrorMessage(e, 'Failed to reject the order');
      await loadOrders(); // the customer may have cancelled it meanwhile
      _error = message;
      notifyListeners();
      return false;
    }
  }

  void _connectWebSocket() {
    final token = _authService.accessToken;
    final sid = shopId;
    if (token == null || sid == null) return;

    _wsClient.addHandler(_handleWSMessage);
    _wsClient.connect(token: token, shopId: sid);
  }

  void _handleWSMessage(Map<String, dynamic> message) {
    final type = message['type'] as String?;
    if (type == 'NEW_ORDER') {
      // Reload orders to get the full order data
      loadOrders();
    } else if (type == 'ORDER_CANCELLED') {
      final data = message['data'] as Map<String, dynamic>?;
      final orderId = data?['orderId'] as String?;
      final idx = orderId == null ? -1 : _orders.indexWhere((o) => o.id == orderId);
      if (idx >= 0) {
        _orders[idx].status = OrderStatus.cancelled;
        final code = _orders[idx].pickupCode;
        _notices.add(code.isNotEmpty
            ? 'Order $code was cancelled by the customer'
            : 'An order was cancelled by the customer');
        notifyListeners();
      }
    } else if (type == 'ORDER_STATUS_CHANGED') {
      final data = message['data'] as Map<String, dynamic>?;
      if (data != null) {
        final orderId = data['orderId'] as String?;
        final newStatus = data['newStatus'] as String?;
        if (orderId != null && newStatus != null) {
          final idx = _orders.indexWhere((o) => o.id == orderId);
          if (idx >= 0) {
            _orders[idx].status = OrderStatus.fromString(newStatus);
            notifyListeners();
          }
        }
      }
    }
  }

  @override
  void dispose() {
    _wsClient.removeHandler(_handleWSMessage);
    _notices.close();
    super.dispose();
  }
}
