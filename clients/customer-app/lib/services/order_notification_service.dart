import 'dart:async';
import 'dart:convert';
import 'package:flutter/foundation.dart';
import 'package:flutter_local_notifications/flutter_local_notifications.dart';
import 'package:quickeats_customer/core/constants/order_status.dart';
import 'package:quickeats_customer/core/network/ws_client.dart';
import 'package:quickeats_customer/services/auth_service.dart';
import 'package:shared_preferences/shared_preferences.dart';

/// A live change to one of the customer's orders.
class OrderUpdate {
  final String orderId;
  final OrderStatus status;
  final String? pickupCode;
  final String? reason;
  const OrderUpdate({required this.orderId, required this.status, this.pickupCode, this.reason});
}

/// Keeps a WebSocket open to the notification service while a customer is logged in,
/// turns each order update into a phone notification, and re-broadcasts it so open
/// screens can refresh straight away instead of waiting for their next poll.
class OrderNotificationService {
  final AuthService _auth;
  final WSClient _ws;
  final _plugin = FlutterLocalNotificationsPlugin();
  final _updates = StreamController<OrderUpdate>.broadcast();
  bool _connected = false;
  bool _alertsEnabled = true;
  static const _alertsKey = 'order_alerts_enabled';

  OrderNotificationService({required AuthService authService, required WSClient wsClient})
      : _auth = authService, _ws = wsClient;

  Stream<OrderUpdate> get updates => _updates.stream;

  /// Whether status changes pop up as phone notifications (live in-app updates always run).
  bool get alertsEnabled => _alertsEnabled;

  Future<void> setAlertsEnabled(bool enabled) async {
    _alertsEnabled = enabled;
    try {
      await (await SharedPreferences.getInstance()).setBool(_alertsKey, enabled);
    } catch (e) {
      debugPrint('[Notify] could not save preference: $e');
    }
    if (enabled) await requestPermission();
  }

  static const _channel = AndroidNotificationDetails(
    'order_updates', 'Order updates',
    channelDescription: 'Status changes for your QuickEats orders',
    importance: Importance.high,
    priority: Priority.high,
  );

  Future<void> init() async {
    try {
      _alertsEnabled = (await SharedPreferences.getInstance()).getBool(_alertsKey) ?? true;
    } catch (_) {}
    if (!kIsWeb) {
      await _plugin.initialize(
        settings: const InitializationSettings(android: AndroidInitializationSettings('@mipmap/ic_launcher')),
      );
    }
    _ws.addHandler(_onMessage);
    _auth.addListener(_syncConnection);
    _syncConnection();
  }

  /// Asks for the Android 13+ notification permission. Call once a screen is showing.
  Future<void> requestPermission() async {
    if (kIsWeb || !_alertsEnabled) return;
    await _plugin
        .resolvePlatformSpecificImplementation<AndroidFlutterLocalNotificationsPlugin>()
        ?.requestNotificationsPermission();
  }

  // Connect on login, disconnect on logout.
  void _syncConnection() {
    final isCustomer = _auth.isAuthenticated && (_auth.currentUser?.role ?? 'customer') == 'customer';
    if (isCustomer && !_connected) {
      _connected = true;
      _ws.connect(_freshToken);
    } else if (!isCustomer && _connected) {
      _connected = false;
      _ws.disconnect();
    }
  }

  // The server checks the token only when the socket opens, so refresh it first if it
  // is about to expire (access tokens last 15 minutes).
  Future<String?> _freshToken() async {
    final token = _auth.accessToken;
    if (token == null) return null;
    if (_expiresWithin(token, const Duration(minutes: 1))) {
      if (!await _auth.refreshTokens()) return null;
    }
    return _auth.accessToken;
  }

  bool _expiresWithin(String jwt, Duration margin) {
    try {
      final payload = jsonDecode(utf8.decode(base64Url.decode(base64Url.normalize(jwt.split('.')[1]))));
      final exp = DateTime.fromMillisecondsSinceEpoch((payload['exp'] as int) * 1000);
      return DateTime.now().add(margin).isAfter(exp);
    } catch (_) {
      return false;
    }
  }

  void _onMessage(Map<String, dynamic> message) {
    if (message['type'] != 'ORDER_STATUS_CHANGED') return;
    final data = message['data'] as Map<String, dynamic>?;
    final orderId = data?['orderId'] as String?;
    final status = data?['newStatus'] as String?;
    if (orderId == null || status == null) return;

    final update = OrderUpdate(
      orderId: orderId,
      status: OrderStatus.fromString(status),
      pickupCode: data?['pickupCode'] as String?,
      reason: data?['reason'] as String?,
    );
    // The "placed" event travels on its own topic and can arrive after the order has
    // already moved on, so it only raises an alert and never rolls back an open screen.
    if (update.status != OrderStatus.pendingPayment) _updates.add(update);

    // No alert for a cancellation the customer made themselves.
    if (update.reason == 'cancelled_by_customer') return;
    _showNotification(update);
  }

  Future<void> _showNotification(OrderUpdate update) async {
    if (kIsWeb || !_alertsEnabled) return;
    final placed = update.status == OrderStatus.pendingPayment;
    var body = placed
        ? 'We received your order and are processing your payment.'
        : update.status.customerMessage;
    if (update.status == OrderStatus.confirmed && (update.pickupCode ?? '').isNotEmpty) {
      body = '$body Pickup code: ${update.pickupCode}';
    }
    try {
      await _plugin.show(
        // One notification per order: a newer status replaces the older one.
        id: update.orderId.hashCode & 0x7fffffff,
        title: placed ? 'Order placed' : 'Order ${update.status.label}',
        body: body,
        notificationDetails: const NotificationDetails(android: _channel),
        payload: update.orderId,
      );
    } catch (e) {
      debugPrint('[Notify] show failed: $e');
    }
  }
}
