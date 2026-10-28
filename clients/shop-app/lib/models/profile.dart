import 'package:quickeats_shop/models/organisation.dart';

class Profile {
  final String id;
  final String email;
  final String displayName;
  final String role;
  final String? shopId;
  /// Organisations whose customers can see this merchant's shop.
  final List<Organisation> shopOrganisations;

  const Profile({required this.id, required this.email, required this.displayName, required this.role, this.shopId, this.shopOrganisations = const []});

  factory Profile.fromJson(Map<String, dynamic> json) => Profile(
    id: json['id'] ?? '', email: json['email'] ?? '', displayName: json['displayName'] ?? '', role: json['role'] ?? '',
    shopId: json['shopId'],
    shopOrganisations: (json['shopOrganisations'] as List?)?.map((o) => Organisation.fromJson(o)).toList() ?? const [],
  );

  String get initials {
    final parts = displayName.trim().split(RegExp(r'\s+')).where((p) => p.isNotEmpty).toList();
    if (parts.isEmpty) return '?';
    return (parts.first[0] + (parts.length > 1 ? parts.last[0] : '')).toUpperCase();
  }
}
