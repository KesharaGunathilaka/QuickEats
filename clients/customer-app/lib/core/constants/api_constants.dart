import 'package:flutter/foundation.dart' show kIsWeb;

class ApiConstants {
  // Web runs on the host itself; the Android emulator reaches the host via 10.0.2.2.
  // Change to production URL when deployed.
  static const String baseUrl = kIsWeb ? 'http://localhost:8000' : 'http://10.0.2.2:8000';
  // static const String baseUrl = 'https://api.quickeats.example.com';

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
