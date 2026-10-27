import 'package:flutter/foundation.dart' show kIsWeb;

class ApiConstants {
  // Local dev default: web runs on the host itself; the Android emulator reaches the host via 10.0.2.2.
  // Point at the deployed API without editing code:
  //   flutter run --dart-define=API_BASE_URL=http://13.63.249.122
  static const String baseUrl = String.fromEnvironment(
    'API_BASE_URL',
    defaultValue: kIsWeb ? 'http://localhost:8000' : 'http://10.0.2.2:8000',
  );
  // Live order updates come over a WebSocket through the same Kong gateway.
  static final String wsUrl = '${baseUrl.replaceFirst('http', 'ws')}/ws';

  // Auth
  static const String login = '/api/v1/auth/login';
  static const String register = '/api/v1/auth/register';
  static const String refresh = '/api/v1/auth/refresh';

  // Catalog
  static const String shops = '/api/v1/shops';
  static String shopById(String id) => '/api/v1/shops/$id';
  static String shopMenu(String shopId) => '/api/v1/shops/$shopId/menu';

  // Orders
  static const String orders = '/api/v1/orders';
  static String orderById(String id) => '/api/v1/orders/$id';
  static String cancelOrder(String id) => '/api/v1/orders/$id/cancel';

  // Wallet
  static const String wallet = '/api/v1/wallet';
  static const String walletTopup = '/api/v1/wallet/topup';

  // Payments
  static const String paymentHistory = '/api/v1/payments/history';
}
