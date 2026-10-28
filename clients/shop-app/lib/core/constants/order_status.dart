enum OrderStatus {
  pendingPayment('PENDING_PAYMENT', 'Pending Payment'),
  // Paid by the customer and waiting for the shop to accept or reject it.
  confirmed('CONFIRMED', 'New'),
  accepted('ACCEPTED', 'Accepted'),
  // Older workflow steps, no longer set by the app; shown with accepted orders.
  preparing('PREPARING', 'Preparing'),
  ready('READY', 'Ready'),
  pickedUp('PICKED_UP', 'Completed'),
  cancelled('CANCELLED', 'Cancelled'),
  rejected('REJECTED', 'Rejected'),
  refunded('REFUNDED', 'Refunded');

  final String value;
  final String label;
  const OrderStatus(this.value, this.label);

  static OrderStatus fromString(String s) =>
    OrderStatus.values.firstWhere((e) => e.value == s, orElse: () => OrderStatus.pendingPayment);

  bool get isTerminal => this == pickedUp || this == refunded;
  bool get isActive => !isTerminal && this != cancelled && this != rejected;

  /// The queue tab this order is listed under.
  OrderStatus get queueTab => (this == preparing || this == ready) ? accepted : this;

  /// Only a new order can be rejected; once accepted the shop must complete it.
  bool get canReject => this == confirmed;

  OrderStatus? get nextShopAction {
    switch (this) {
      case confirmed: return accepted;
      case accepted:
      case preparing:
      case ready: return pickedUp;
      default: return null;
    }
  }

  String get nextActionLabel {
    switch (this) {
      case confirmed: return 'Accept Order';
      case accepted:
      case preparing:
      case ready: return 'Complete Order';
      default: return '';
    }
  }
}
