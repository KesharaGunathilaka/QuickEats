import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:provider/provider.dart';
import 'package:quickeats_customer/core/constants/app_colors.dart';
import 'package:quickeats_customer/models/payment_card.dart';
import 'package:quickeats_customer/viewmodels/card_viewmodel.dart';
import 'package:quickeats_customer/viewmodels/cart_viewmodel.dart';
import 'package:quickeats_customer/viewmodels/order_viewmodel.dart';

class CheckoutScreen extends StatefulWidget {
  const CheckoutScreen({super.key});
  @override State<CheckoutScreen> createState() => _CheckoutScreenState();
}

class _CheckoutScreenState extends State<CheckoutScreen> {
  int? _checkedTotal;

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) => _checkPayment());
  }

  // Asks the server whether the saved card covers the cart total.
  void _checkPayment() {
    if (!mounted) return;
    final total = context.read<CartViewModel>().totalCents;
    _checkedTotal = total;
    context.read<CardViewModel>().checkPayment(total);
  }

  Future<void> _addCard() async {
    final added = await context.push<bool>('/card/add');
    if (added == true) _checkPayment();
  }

  Future<void> _placeOrder(CartViewModel cart, OrderViewModel orderVM, CardViewModel cardVM) async {
    final order = await orderVM.placeOrder(shopId: cart.shopId!, items: cart.items);
    if (!mounted) return;
    if (order != null) { cart.clear(); context.go('/orders/${order.id}'); return; }
    // The server refused for a payment reason: show the matching banner.
    if (orderVM.errorCode == 'NO_CARD' || orderVM.errorCode == 'INSUFFICIENT_FUNDS') cardVM.markDeclined(orderVM.errorCode!);
  }

  @override
  Widget build(BuildContext context) {
    final cart = context.watch<CartViewModel>();
    final orderVM = context.watch<OrderViewModel>();
    final cardVM = context.watch<CardViewModel>();

    // The cart can change while checkout is open (e.g. back from editing it).
    if (_checkedTotal != null && _checkedTotal != cart.totalCents) {
      WidgetsBinding.instance.addPostFrameCallback((_) => _checkPayment());
      _checkedTotal = cart.totalCents;
    }

    final readiness = cardVM.readiness;
    final canPlace = !orderVM.isPlacing && !cart.isEmpty &&
        (readiness == PaymentReadiness.ready || readiness == PaymentReadiness.unknown);
    final paymentError = orderVM.errorCode == 'NO_CARD' || orderVM.errorCode == 'INSUFFICIENT_FUNDS';

    return Scaffold(
      appBar: AppBar(title: const Text('Checkout', style: TextStyle(fontWeight: FontWeight.w700)), backgroundColor: AppColors.white, elevation: 0),
      body: SingleChildScrollView(padding: const EdgeInsets.all(16), child: Column(crossAxisAlignment: CrossAxisAlignment.stretch, children: [
        // Order summary
        Container(padding: const EdgeInsets.all(16), decoration: BoxDecoration(color: AppColors.white, borderRadius: BorderRadius.circular(12), border: Border.all(color: AppColors.border)),
          child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
            const Text('Order Summary', style: TextStyle(fontSize: 16, fontWeight: FontWeight.w600)),
            const SizedBox(height: 12),
            ...cart.items.map((item) => Padding(padding: const EdgeInsets.only(bottom: 6), child: Row(children: [
              Text('${item.quantity}×', style: const TextStyle(fontWeight: FontWeight.w600, color: AppColors.textSecondary)),
              const SizedBox(width: 8),
              Expanded(child: Text(item.menuItem.name)),
              Text(item.formattedSubtotal, style: const TextStyle(color: AppColors.textSecondary)),
            ]))),
            const Divider(),
            Row(children: [const Text('Total', style: TextStyle(fontSize: 16, fontWeight: FontWeight.w700)), const Spacer(), Text(cart.formattedTotal, style: const TextStyle(fontSize: 18, fontWeight: FontWeight.w700, color: AppColors.primary))]),
          ]),
        ),
        const SizedBox(height: 20),

        // Payment
        const Text('Payment', style: TextStyle(fontSize: 16, fontWeight: FontWeight.w600)),
        const SizedBox(height: 10),
        _PaymentSection(
          readiness: readiness,
          card: cardVM.card,
          total: cart.formattedTotal,
          onAddCard: _addCard,
          onEditCart: () => context.go('/cart'),
          onRetry: _checkPayment,
        ),
        const SizedBox(height: 24),

        if (orderVM.error != null && !paymentError)
          Padding(padding: const EdgeInsets.only(bottom: 12), child: Text(orderVM.error!, style: const TextStyle(color: AppColors.error))),

        SizedBox(height: 54, child: FilledButton(
          onPressed: canPlace ? () => _placeOrder(cart, orderVM, cardVM) : null,
          style: FilledButton.styleFrom(backgroundColor: AppColors.primary, shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12))),
          child: orderVM.isPlacing
            ? const SizedBox(height: 22, width: 22, child: CircularProgressIndicator(strokeWidth: 2, color: Colors.white))
            : Text(_buttonLabel(readiness, cart.formattedTotal), style: const TextStyle(fontSize: 16, fontWeight: FontWeight.w600)),
        )),
      ])),
    );
  }

  String _buttonLabel(PaymentReadiness r, String total) {
    switch (r) {
      case PaymentReadiness.noCard: return 'Add a card to place your order';
      case PaymentReadiness.insufficient: return 'Insufficient balance';
      case PaymentReadiness.loading: return 'Checking your card...';
      default: return 'Pay $total & Place Order';
    }
  }
}

/// The payment block: the saved card when it can pay, otherwise what's wrong and
/// the button that fixes it.
class _PaymentSection extends StatelessWidget {
  final PaymentReadiness readiness;
  final PaymentCard? card;
  final String total;
  final VoidCallback onAddCard;
  final VoidCallback onEditCart;
  final VoidCallback onRetry;

  const _PaymentSection({required this.readiness, required this.card, required this.total, required this.onAddCard, required this.onEditCart, required this.onRetry});

  @override
  Widget build(BuildContext context) {
    switch (readiness) {
      case PaymentReadiness.loading:
        return _box(child: const Row(children: [
          SizedBox(width: 20, height: 20, child: CircularProgressIndicator(strokeWidth: 2)),
          SizedBox(width: 12),
          Text('Checking your card...', style: TextStyle(color: AppColors.textSecondary)),
        ]));

      case PaymentReadiness.noCard:
        return _Notice(
          color: AppColors.warning,
          icon: Icons.credit_card_off_outlined,
          title: 'No card added',
          message: 'You need a payment card to place this order. Add one and you\'ll come straight back here.',
          actions: [
            FilledButton.icon(onPressed: onAddCard, icon: const Icon(Icons.add_card, size: 18), label: const Text('Add a card'),
                style: FilledButton.styleFrom(backgroundColor: AppColors.primary)),
          ],
        );

      case PaymentReadiness.insufficient:
        return _Notice(
          color: AppColors.error,
          icon: Icons.money_off_csred_outlined,
          title: 'Insufficient balance',
          message: 'Your ${card?.shortLabel ?? 'card'} doesn\'t have enough balance for $total. Use a different card or remove some items.',
          actions: [
            FilledButton.icon(onPressed: onAddCard, icon: const Icon(Icons.swap_horiz, size: 18), label: const Text('Use another card'),
                style: FilledButton.styleFrom(backgroundColor: AppColors.primary)),
            OutlinedButton(onPressed: onEditCart, child: const Text('Edit cart')),
          ],
        );

      case PaymentReadiness.unknown:
        return _Notice(
          color: AppColors.textSecondary,
          icon: Icons.cloud_off,
          title: 'Couldn\'t check your card',
          message: 'You can still place the order; payment is checked when it goes through.',
          actions: [OutlinedButton(onPressed: onRetry, child: const Text('Check again'))],
        );

      case PaymentReadiness.ready:
        return _box(child: Row(children: [
          Container(
            padding: const EdgeInsets.all(10),
            decoration: BoxDecoration(color: AppColors.primaryLight, borderRadius: BorderRadius.circular(10)),
            child: const Icon(Icons.credit_card, color: AppColors.primary),
          ),
          const SizedBox(width: 12),
          Expanded(child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
            Text(card?.shortLabel ?? 'Card', style: const TextStyle(fontWeight: FontWeight.w600)),
            const SizedBox(height: 2),
            Text('Expires ${card?.expiry ?? ''}', style: const TextStyle(fontSize: 12, color: AppColors.textSecondary)),
          ])),
          const Icon(Icons.check_circle, color: AppColors.secondary, size: 20),
          TextButton(onPressed: onAddCard, child: const Text('Change')),
        ]));
    }
  }

  Widget _box({required Widget child}) => Container(
    padding: const EdgeInsets.all(14),
    decoration: BoxDecoration(color: AppColors.white, borderRadius: BorderRadius.circular(12), border: Border.all(color: AppColors.border)),
    child: child,
  );
}

/// A tinted banner with an icon, a title, a message and action buttons.
class _Notice extends StatelessWidget {
  final Color color;
  final IconData icon;
  final String title;
  final String message;
  final List<Widget> actions;
  const _Notice({required this.color, required this.icon, required this.title, required this.message, required this.actions});

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(color: color.withOpacity(0.08), borderRadius: BorderRadius.circular(12), border: Border.all(color: color.withOpacity(0.4))),
      child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
        Row(children: [
          Icon(icon, color: color),
          const SizedBox(width: 10),
          Expanded(child: Text(title, style: TextStyle(fontSize: 15, fontWeight: FontWeight.w700, color: color))),
        ]),
        const SizedBox(height: 8),
        Text(message, style: const TextStyle(color: AppColors.textPrimary, height: 1.35)),
        const SizedBox(height: 12),
        Wrap(spacing: 10, runSpacing: 8, children: actions),
      ]),
    );
  }
}
