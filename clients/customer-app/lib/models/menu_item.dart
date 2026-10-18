class MenuItem {
  final String id;
  final String shopId;
  final String name;
  final String description;
  final int price;
  final String category;
  final int stock;
  final bool available;

  MenuItem({required this.id, required this.shopId, required this.name, required this.description, required this.price, required this.category, required this.stock, required this.available});

  factory MenuItem.fromJson(Map<String, dynamic> json) => MenuItem(
    id: json['id'] ?? '', shopId: json['shopId'] ?? '', name: json['name'] ?? '',
    description: json['description'] ?? '', price: json['price'] ?? 0,
    category: json['category'] ?? '', stock: json['stock'] ?? 0, available: json['available'] ?? false,
  );

  String get formattedPrice => 'LKR ${(price / 100).toStringAsFixed(2)}';
}
