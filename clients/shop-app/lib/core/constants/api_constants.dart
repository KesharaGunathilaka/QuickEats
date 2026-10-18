import 'package:flutter/foundation.dart' show kIsWeb;

class ApiConstants {
  // Web runs on the host itself; the Android emulator reaches the host via 10.0.2.2.
  static const String _host = kIsWeb ? 'localhost' : '10.0.2.2';
  static const String baseUrl = 'http://$_host:8000';
  static const String wsUrl = 'ws://$_host:8086/ws';

  static const String login = '/api/v1/auth/login';
  static const String register = '/api/v1/auth/register';
  static const String registerShop = '/api/v1/auth/register-shop';
  static const String refresh = '/api/v1/auth/refresh';

  static const String shops = '/api/v1/shops';
  static String shopMenu(String shopId) => '/api/v1/shops/$shopId/menu';
  static String createMenuItem(String shopId) => '/api/v1/shops/$shopId/menu-items';
  static String menuItem(String itemId) => '/api/v1/menu-items/$itemId';
  static String updateStock(String itemId) => '/api/v1/menu-items/$itemId/stock';

  static String shopOrders(String shopId) => '/api/v1/shops/$shopId/orders';
  static String orderById(String id) => '/api/v1/orders/$id';
  static String updateOrderStatus(String id) => '/api/v1/orders/$id/status';
  static String rejectOrder(String id) => '/api/v1/orders/$id/reject';

  static String shopSettlements(String shopId) => '/api/v1/shops/$shopId/settlements';
}
