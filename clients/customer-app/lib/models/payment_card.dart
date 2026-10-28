/// The customer's saved payment card. The server only ever sends display details;
/// there is no balance here by design.
class PaymentCard {
  final String id;
  final String brand;
  final String last4;
  final String holderName;
  final int expMonth;
  final int expYear;

  PaymentCard({required this.id, required this.brand, required this.last4, required this.holderName, required this.expMonth, required this.expYear});

  factory PaymentCard.fromJson(Map<String, dynamic> json) => PaymentCard(
    id: json['id'] ?? '', brand: json['brand'] ?? 'Card', last4: json['last4'] ?? '',
    holderName: json['holderName'] ?? '', expMonth: json['expMonth'] ?? 0, expYear: json['expYear'] ?? 0,
  );

  String get maskedNumber => '•••• •••• •••• $last4';
  String get shortLabel => '$brand •••• $last4';
  String get expiry => '${expMonth.toString().padLeft(2, '0')}/${(expYear % 100).toString().padLeft(2, '0')}';
}

/// Whether the saved card can pay a given amount.
class PayCheck {
  final bool hasCard;
  final bool sufficient;
  const PayCheck({required this.hasCard, required this.sufficient});

  factory PayCheck.fromJson(Map<String, dynamic> json) =>
      PayCheck(hasCard: json['hasCard'] == true, sufficient: json['sufficient'] == true);
}
