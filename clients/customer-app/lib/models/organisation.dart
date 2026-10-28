class Organisation {
  final String id;
  final String name;
  final String slug;
  final List<String> emailDomains;

  const Organisation({required this.id, required this.name, required this.slug, required this.emailDomains});

  factory Organisation.fromJson(Map<String, dynamic> json) => Organisation(
    id: json['id'] ?? '', name: json['name'] ?? '', slug: json['slug'] ?? '',
    emailDomains: (json['emailDomains'] as List?)?.cast<String>() ?? const [],
  );

  /// Same rule as the server: exact domain, or "*.ruh.ac.lk" for any subdomain.
  bool acceptsEmail(String email) {
    final at = email.lastIndexOf('@');
    if (at < 1 || at == email.length - 1) return false;
    final domain = email.substring(at + 1).toLowerCase();
    return emailDomains.any((d) => d.startsWith('*.') ? domain.endsWith(d.substring(1)) : domain == d);
  }

  /// "@eng.ruh.ac.lk or @*.ruh.ac.lk"
  String get domainsLabel => emailDomains.map((d) => '@$d').join(' or ');
}
