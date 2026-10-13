class Shop {
  final String id;
  final String organisationId;
  final String name;
  final String description;
  final String imageUrl;
  final bool isOpen;
  final String ownerId;

  Shop({required this.id, required this.organisationId, required this.name, required this.description, required this.imageUrl, required this.isOpen, required this.ownerId});

  factory Shop.fromJson(Map<String, dynamic> json) => Shop(
    id: json['id'] ?? '',
    organisationId: json['organisationId'] ?? '',
    name: json['name'] ?? '',
    description: json['description'] ?? '',
    imageUrl: json['imageUrl'] ?? '',
    isOpen: json['isOpen'] ?? false,
    ownerId: json['ownerId'] ?? '',
  );
}
