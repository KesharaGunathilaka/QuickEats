import 'package:quickeats_customer/models/organisation.dart';

class Profile {
  final String id;
  final String email;
  final String displayName;
  final String role;
  final Organisation? organisation;

  const Profile({required this.id, required this.email, required this.displayName, required this.role, this.organisation});

  factory Profile.fromJson(Map<String, dynamic> json) => Profile(
    id: json['id'] ?? '', email: json['email'] ?? '', displayName: json['displayName'] ?? '', role: json['role'] ?? '',
    organisation: json['organisation'] is Map<String, dynamic> ? Organisation.fromJson(json['organisation']) : null,
  );

  String get initials {
    final parts = displayName.trim().split(RegExp(r'\s+')).where((p) => p.isNotEmpty).toList();
    if (parts.isEmpty) return '?';
    return (parts.first[0] + (parts.length > 1 ? parts.last[0] : '')).toUpperCase();
  }
}
