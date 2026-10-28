import 'package:flutter/material.dart';
import 'package:quickeats_customer/core/constants/app_colors.dart';
import 'package:quickeats_customer/models/payment_card.dart';

/// A payment card drawn as a card. Shows brand, masked number, holder and expiry,
/// never a balance. With [card] null it shows a preview from the typed-in values.
class CardVisual extends StatelessWidget {
  final PaymentCard? card;
  final String? brand;
  final String? number;
  final String? holder;
  final String? expiry;

  const CardVisual({super.key, this.card, this.brand, this.number, this.holder, this.expiry});

  @override
  Widget build(BuildContext context) {
    final b = card?.brand ?? brand ?? 'Card';
    final num = card?.maskedNumber ?? (number == null || number!.isEmpty ? '•••• •••• •••• ••••' : number!);
    final name = card?.holderName ?? (holder == null || holder!.isEmpty ? 'CARDHOLDER NAME' : holder!);
    final exp = card?.expiry ?? (expiry == null || expiry!.isEmpty ? 'MM/YY' : expiry!);

    return AspectRatio(
      aspectRatio: 1.586, // standard card proportions
      child: Container(
        padding: const EdgeInsets.all(22),
        decoration: BoxDecoration(
          gradient: LinearGradient(colors: _colors(b), begin: Alignment.topLeft, end: Alignment.bottomRight),
          borderRadius: BorderRadius.circular(18),
          boxShadow: [BoxShadow(color: _colors(b).last.withOpacity(0.35), blurRadius: 16, offset: const Offset(0, 8))],
        ),
        child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
          Row(children: [
            const Icon(Icons.contactless_outlined, color: Colors.white70, size: 26),
            const Spacer(),
            Text(b.toUpperCase(), style: const TextStyle(color: Colors.white, fontSize: 18, fontWeight: FontWeight.w800, fontStyle: FontStyle.italic, letterSpacing: 1)),
          ]),
          const Spacer(),
          FittedBox(fit: BoxFit.scaleDown, child: Text(num, style: const TextStyle(color: Colors.white, fontSize: 22, fontWeight: FontWeight.w600, letterSpacing: 2))),
          const SizedBox(height: 18),
          Row(crossAxisAlignment: CrossAxisAlignment.end, children: [
            Expanded(child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
              const Text('CARDHOLDER', style: TextStyle(color: Colors.white60, fontSize: 10, letterSpacing: 1)),
              const SizedBox(height: 2),
              Text(name.toUpperCase(), maxLines: 1, overflow: TextOverflow.ellipsis, style: const TextStyle(color: Colors.white, fontSize: 14, fontWeight: FontWeight.w600)),
            ])),
            Column(crossAxisAlignment: CrossAxisAlignment.end, children: [
              const Text('EXPIRES', style: TextStyle(color: Colors.white60, fontSize: 10, letterSpacing: 1)),
              const SizedBox(height: 2),
              Text(exp, style: const TextStyle(color: Colors.white, fontSize: 14, fontWeight: FontWeight.w600)),
            ]),
          ]),
        ]),
      ),
    );
  }

  static List<Color> _colors(String brand) {
    switch (brand) {
      case 'Visa': return const [AppColors.primary, Color(0xFF1E3A8A)];
      case 'Mastercard': return const [Color(0xFF334155), Color(0xFF0F172A)];
      case 'Amex': return const [Color(0xFF0E7490), Color(0xFF164E63)];
      default: return const [Color(0xFF475569), Color(0xFF1E293B)];
    }
  }
}

/// Brand from the first digits as the customer types (matches the server's rules).
String detectCardBrand(String digits) {
  if (digits.startsWith('4')) return 'Visa';
  if (digits.startsWith('34') || digits.startsWith('37')) return 'Amex';
  if (digits.length >= 2) {
    final two = int.tryParse(digits.substring(0, 2)) ?? 0;
    if (two >= 51 && two <= 55) return 'Mastercard';
  }
  if (digits.length >= 4) {
    final four = int.tryParse(digits.substring(0, 4)) ?? 0;
    if (four >= 2221 && four <= 2720) return 'Mastercard';
  }
  return 'Card';
}
