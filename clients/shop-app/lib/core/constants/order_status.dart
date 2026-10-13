enum OrderStatus {
  pendingPayment('PENDING_PAYMENT', 'Pending Payment'),
  confirmed('CONFIRMED', 'Confirmed'),
  accepted('ACCEPTED', 'Accepted'),
  preparing('PREPARING', 'Preparing'),
  ready('READY', 'Ready'),
  pickedUp('PICKED_UP', 'Picked Up'),
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

  OrderStatus? get nextShopAction {
    switch (this) {
      case confirmed: return accepted;
      case accepted: return preparing;
      case preparing: return ready;
      case ready: return pickedUp;
      default: return null;
    }
  }

  String get nextActionLabel {
    switch (this) {
      case confirmed: return 'Accept Order';
      case accepted: return 'Start Preparing';
      case preparing: return 'Mark Ready';
      case ready: return 'Confirm Pickup';
      default: return '';
    }
  }
}
