class Wallet {
  final String id;
  final int balance;
  final String currency;

  Wallet({required this.id, required this.balance, required this.currency});

  factory Wallet.fromJson(Map<String, dynamic> json) => Wallet(
    id: json['id'] ?? '', balance: json['balance'] ?? 0, currency: json['currency'] ?? 'LKR',
  );

  String get formattedBalance => 'LKR ${(balance / 100).toStringAsFixed(2)}';
}
