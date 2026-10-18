class Shop {
  final String id;
  final String name;
  final String description;
  final String imageUrl;
  final bool isOpen;

  Shop({required this.id, required this.name, required this.description, required this.imageUrl, required this.isOpen});

  factory Shop.fromJson(Map<String, dynamic> json) => Shop(
    id: json['id'] ?? '', name: json['name'] ?? '', description: json['description'] ?? '',
    imageUrl: json['imageUrl'] ?? '', isOpen: json['isOpen'] ?? false,
  );
}
