import 'package:flutter/material.dart';

enum OrderStatus {
  pendingPayment('PENDING_PAYMENT', 'Pending Payment', Icons.hourglass_top, Color(0xFFF59E0B)),
  // Paid and waiting for the shop to accept or reject it.
  confirmed('CONFIRMED', 'Awaiting Shop', Icons.storefront_outlined, Color(0xFFF59E0B)),
  accepted('ACCEPTED', 'Accepted', Icons.thumb_up_outlined, Color(0xFF8B5CF6)),
  preparing('PREPARING', 'Preparing', Icons.restaurant, Color(0xFFF97316)),
  ready('READY', 'Ready for Pickup!', Icons.takeout_dining, Color(0xFF22C55E)),
  pickedUp('PICKED_UP', 'Picked Up', Icons.check_circle, Color(0xFF6B7280)),
  cancelled('CANCELLED', 'Cancelled', Icons.cancel_outlined, Color(0xFFEF4444)),
  rejected('REJECTED', 'Rejected', Icons.block, Color(0xFFEF4444)),
  refunded('REFUNDED', 'Refunded', Icons.replay, Color(0xFF6B7280));

  final String value;
  final String label;
  final IconData icon;
  final Color color;
  const OrderStatus(this.value, this.label, this.icon, this.color);

  static OrderStatus fromString(String s) =>
    OrderStatus.values.firstWhere((e) => e.value == s, orElse: () => OrderStatus.pendingPayment);

  bool get isTerminal => this == pickedUp || this == refunded;
  bool get isActive => !isTerminal && this != cancelled && this != rejected;
  /// The customer can cancel until the shop accepts the order.
  bool get canCancel => this == pendingPayment || this == confirmed;

  String get customerMessage {
    switch (this) {
      case pendingPayment: return 'Completing your payment...';
      case confirmed: return 'Waiting for the shop to accept your order. You can still cancel it until they do.';
      case accepted: return 'The shop accepted your order! Show your pickup code when you collect it.';
      case preparing: return 'Your food is being prepared 🍳';
      case ready: return 'Your order is ready! Head to the shop to pick it up.';
      case pickedUp: return 'Enjoy your meal! 😋';
      case cancelled: return 'This order was cancelled.';
      case rejected: return 'Sorry, the shop couldn\'t fulfil this order.';
      case refunded: return 'Your refund has been processed.';
    }
  }
}
