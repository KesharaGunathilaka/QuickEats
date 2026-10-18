class User {
  final String id;
  final String organisationId;
  final String email;
  final String displayName;
  final String role;
  final String? shopId;

  User({required this.id, required this.organisationId, required this.email, required this.displayName, required this.role, this.shopId});

  factory User.fromJson(Map<String, dynamic> json) => User(
    id: json['id'] ?? '', organisationId: json['organisationId'] ?? '',
    email: json['email'] ?? '', displayName: json['displayName'] ?? '',
    role: json['role'] ?? '', shopId: json['shopId'],
  );

  Map<String, dynamic> toJson() => {'id': id, 'organisationId': organisationId, 'email': email, 'displayName': displayName, 'role': role, 'shopId': shopId};
}
