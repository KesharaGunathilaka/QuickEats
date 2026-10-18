import 'package:quickeats_shop/core/constants/order_status.dart';

class Order {
  final String id;
  final String customerId;
  final String shopId;
  OrderStatus status;
  final String pickupCode;
  final int totalAmount; // cents
  final String currency;
  final String paymentMethod;
  final List<OrderItem> items;
  final String rejectReason;
  final DateTime createdAt;
  final DateTime updatedAt;

  Order({
    required this.id, required this.customerId, required this.shopId,
    required this.status, required this.pickupCode, required this.totalAmount,
    required this.currency, required this.paymentMethod, required this.items,
    required this.rejectReason, required this.createdAt, required this.updatedAt,
  });

  factory Order.fromJson(Map<String, dynamic> json) => Order(
    id: json['id'] ?? '',
    customerId: json['customerId'] ?? '',
    shopId: json['shopId'] ?? '',
    status: OrderStatus.fromString(json['status'] ?? 'PENDING_PAYMENT'),
    pickupCode: json['pickupCode'] ?? '',
    totalAmount: json['totalAmount'] ?? 0,
    currency: json['currency'] ?? 'LKR',
    paymentMethod: json['paymentMethod'] ?? '',
    items: (json['items'] as List<dynamic>?)?.map((e) => OrderItem.fromJson(e)).toList() ?? [],
    rejectReason: json['rejectReason'] ?? '',
    createdAt: DateTime.tryParse(json['createdAt'] ?? '') ?? DateTime.now(),
    updatedAt: DateTime.tryParse(json['updatedAt'] ?? '') ?? DateTime.now(),
  );

  String get formattedTotal => 'LKR ${(totalAmount / 100).toStringAsFixed(2)}';
  String get shortId => id.substring(0, 8).toUpperCase();
  Duration get waitTime => DateTime.now().difference(createdAt);
  String get waitTimeFormatted {
    final m = waitTime.inMinutes;
    if (m < 1) return 'Just now';
    if (m < 60) return '${m}m ago';
    return '${m ~/ 60}h ${m % 60}m ago';
  }
}

class OrderItem {
  final String id;
  final String menuItemId;
  final String itemName;
  final int quantity;
  final int unitPrice;
  final int subtotal;

  OrderItem({required this.id, required this.menuItemId, required this.itemName, required this.quantity, required this.unitPrice, required this.subtotal});

  factory OrderItem.fromJson(Map<String, dynamic> json) => OrderItem(
    id: json['id'] ?? '',
    menuItemId: json['menuItemId'] ?? '',
    itemName: json['itemName'] ?? '',
    quantity: json['quantity'] ?? 0,
    unitPrice: json['unitPrice'] ?? 0,
    subtotal: json['subtotal'] ?? 0,
  );
}
