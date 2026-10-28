import 'package:quickeats_customer/core/constants/order_status.dart';

class Order {
  final String id;
  final String shopId;
  OrderStatus status;
  final String pickupCode;
  final int totalAmount;
  final String currency;
  final String paymentMethod;
  final List<OrderItem> items;
  final DateTime createdAt;
  /// Why the order was cancelled: insufficient_funds, no_card, payment_failed,
  /// payment_timeout, stock_insufficient, cancelled_by_customer.
  final String cancelReason;
  final String rejectReason;

  Order({required this.id, required this.shopId, required this.status, required this.pickupCode, required this.totalAmount, required this.currency, required this.paymentMethod, required this.items, required this.createdAt, this.cancelReason = '', this.rejectReason = ''});

  /// The card was declined, so the customer needs to add or change a card.
  bool get paymentDeclined => cancelReason == 'insufficient_funds' || cancelReason == 'no_card' || cancelReason == 'payment_failed';

  factory Order.fromJson(Map<String, dynamic> json) => Order(
    id: json['id'] ?? '', shopId: json['shopId'] ?? '',
    status: OrderStatus.fromString(json['status'] ?? ''),
    pickupCode: json['pickupCode'] ?? '', totalAmount: json['totalAmount'] ?? 0,
    currency: json['currency'] ?? 'LKR', paymentMethod: json['paymentMethod'] ?? '',
    items: (json['items'] as List?)?.map((e) => OrderItem.fromJson(e)).toList() ?? [],
    createdAt: DateTime.tryParse(json['createdAt'] ?? '') ?? DateTime.now(),
    cancelReason: json['cancelReason'] ?? '', rejectReason: json['rejectReason'] ?? '',
  );

  String get formattedTotal => 'LKR ${(totalAmount / 100).toStringAsFixed(2)}';
  String get shortId => id.length >= 8 ? id.substring(0, 8).toUpperCase() : id;
}

class OrderItem {
  final String id;
  final String itemName;
  final int quantity;
  final int unitPrice;
  final int subtotal;

  OrderItem({required this.id, required this.itemName, required this.quantity, required this.unitPrice, required this.subtotal});

  factory OrderItem.fromJson(Map<String, dynamic> json) => OrderItem(
    id: json['id'] ?? '', itemName: json['itemName'] ?? '',
    quantity: json['quantity'] ?? 0, unitPrice: json['unitPrice'] ?? 0, subtotal: json['subtotal'] ?? 0,
  );
}
