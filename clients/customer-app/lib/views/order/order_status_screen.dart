import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:provider/provider.dart';
import 'package:qr_flutter/qr_flutter.dart';
import 'package:quickeats_customer/core/constants/app_colors.dart';
import 'package:quickeats_customer/core/constants/order_status.dart';
import 'package:quickeats_customer/models/order.dart';
import 'package:quickeats_customer/viewmodels/order_viewmodel.dart';

class OrderStatusScreen extends StatefulWidget {
  final String orderId;
  const OrderStatusScreen({super.key, required this.orderId});
  @override State<OrderStatusScreen> createState() => _OrderStatusScreenState();
}

class _OrderStatusScreenState extends State<OrderStatusScreen> {
  // Kept so dispose doesn't look up the provider through a deactivated context.
  late final OrderViewModel _vm;

  @override void initState() { super.initState(); _vm = context.read<OrderViewModel>(); WidgetsBinding.instance.addPostFrameCallback((_) { if (mounted) _vm.trackOrder(widget.orderId); }); }
  @override void dispose() { _vm.stopPolling(); super.dispose(); }

  @override Widget build(BuildContext context) {
    return Consumer<OrderViewModel>(builder: (_, vm, __) {
      final order = vm.activeOrder;
      if (order == null) return Scaffold(appBar: AppBar(), body: const Center(child: CircularProgressIndicator()));
      final color = order.status.color;
      return Scaffold(
        backgroundColor: AppColors.surface,
        appBar: AppBar(title: Text('#${order.shortId}'), backgroundColor: AppColors.white, elevation: 0),
        body: SingleChildScrollView(padding: const EdgeInsets.all(16), child: Column(crossAxisAlignment: CrossAxisAlignment.stretch, children: [
          // Status card
          Container(padding: const EdgeInsets.all(24), decoration: BoxDecoration(color: color.withOpacity(0.08), borderRadius: BorderRadius.circular(16)),
            child: Column(children: [
              Icon(order.status.icon, size: 48, color: color),
              const SizedBox(height: 12),
              Text(order.status.label, style: TextStyle(fontSize: 20, fontWeight: FontWeight.w700, color: color)),
              const SizedBox(height: 6),
              Text(order.status.customerMessage, textAlign: TextAlign.center, style: TextStyle(fontSize: 14, color: AppColors.textSecondary)),
              if (order.status.isActive) ...[
                const SizedBox(height: 16),
                const SizedBox(width: 24, height: 24, child: CircularProgressIndicator(strokeWidth: 2)),
              ],
            ]),
          ),
          const SizedBox(height: 16),

          // Why payment failed, and the way to fix it
          if (order.status == OrderStatus.cancelled && order.paymentDeclined) ...[
            _PaymentDeclinedBanner(order: order),
            const SizedBox(height: 16),
          ],

          // Pickup code + QR
          if (order.pickupCode.isNotEmpty) Container(
            padding: const EdgeInsets.all(24), decoration: BoxDecoration(color: AppColors.white, borderRadius: BorderRadius.circular(16), border: Border.all(color: AppColors.border)),
            child: Column(children: [
              const Text('PICKUP CODE', style: TextStyle(fontSize: 12, fontWeight: FontWeight.w600, color: AppColors.textMuted, letterSpacing: 1.5)),
              const SizedBox(height: 10),
              Text(order.pickupCode, style: TextStyle(fontSize: 40, fontWeight: FontWeight.w800, color: AppColors.primary, letterSpacing: 6)),
              const SizedBox(height: 16),
              QrImageView(data: '${order.id}:${order.pickupCode}', size: 150, backgroundColor: Colors.white),
              const SizedBox(height: 8),
              const Text('Show this to the shop staff', style: TextStyle(fontSize: 12, color: AppColors.textMuted)),
            ]),
          ),
          const SizedBox(height: 16),

          // Items
          Container(padding: const EdgeInsets.all(16), decoration: BoxDecoration(color: AppColors.white, borderRadius: BorderRadius.circular(12), border: Border.all(color: AppColors.border)),
            child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
              const Text('Items', style: TextStyle(fontWeight: FontWeight.w600)),
              const SizedBox(height: 10),
              ...order.items.map((item) => Padding(padding: const EdgeInsets.only(bottom: 4), child: Text('${item.quantity}× ${item.itemName}', style: const TextStyle(fontSize: 14, color: AppColors.textSecondary)))),
              const Divider(),
              Row(children: [const Text('Total', style: TextStyle(fontWeight: FontWeight.w700)), const Spacer(), Text(order.formattedTotal, style: const TextStyle(fontWeight: FontWeight.w700, color: AppColors.primary))]),
            ]),
          ),
          const SizedBox(height: 20),

          // Cancel button
          if (order.status.canCancel) SizedBox(height: 48, child: OutlinedButton(
            onPressed: () async {
              // Pop with the dialog's own context: the screen's context belongs to the
              // shell's nested navigator, so popping it would close the screen, not the dialog.
              final confirm = await showDialog<bool>(context: context, builder: (dialogContext) => AlertDialog(
                title: const Text('Cancel Order?'), content: const Text('Are you sure? Your payment will be refunded.'),
                actions: [TextButton(onPressed: () => Navigator.pop(dialogContext, false), child: const Text('Keep Order')),
                  FilledButton(onPressed: () => Navigator.pop(dialogContext, true), style: FilledButton.styleFrom(backgroundColor: AppColors.error), child: const Text('Cancel Order'))],
              ));
              if (confirm != true || !context.mounted) return;
              final ok = await vm.cancelOrder(order.id);
              if (!ok && context.mounted) {
                ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(vm.error ?? 'Could not cancel the order')));
              }
            },
            style: OutlinedButton.styleFrom(foregroundColor: AppColors.error, side: const BorderSide(color: AppColors.error), shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12))),
            child: const Text('Cancel Order', style: TextStyle(fontWeight: FontWeight.w600)),
          )),
        ])),
      );
    });
  }
}

/// Shown on an order cancelled because the card was declined.
class _PaymentDeclinedBanner extends StatelessWidget {
  final Order order;
  const _PaymentDeclinedBanner({required this.order});

  @override
  Widget build(BuildContext context) {
    final (title, message) = switch (order.cancelReason) {
      'insufficient_funds' => ('Insufficient balance', "Your card didn't have enough balance for ${order.formattedTotal}, so you weren't charged. Add a different card and order again."),
      'no_card' => ('No card on file', 'There was no card to charge, so the order was cancelled. Add a card and order again.'),
      _ => ('Payment declined', "The payment didn't go through, so you weren't charged. Check your card or add a new one."),
    };
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(color: AppColors.error.withOpacity(0.06), borderRadius: BorderRadius.circular(12), border: Border.all(color: AppColors.error.withOpacity(0.35))),
      child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
        Row(children: [
          const Icon(Icons.credit_card_off_outlined, color: AppColors.error),
          const SizedBox(width: 10),
          Expanded(child: Text(title, style: const TextStyle(fontSize: 15, fontWeight: FontWeight.w700, color: AppColors.error))),
        ]),
        const SizedBox(height: 8),
        Text(message, style: const TextStyle(color: AppColors.textPrimary, height: 1.35)),
        const SizedBox(height: 12),
        Wrap(spacing: 10, runSpacing: 8, children: [
          FilledButton.icon(
            onPressed: () => context.push('/card/add'),
            icon: const Icon(Icons.add_card, size: 18),
            label: const Text('Add a new card'),
            style: FilledButton.styleFrom(backgroundColor: AppColors.primary),
          ),
          OutlinedButton(onPressed: () => context.go('/shops'), child: const Text('Back to shops')),
        ]),
      ]),
    );
  }
}
