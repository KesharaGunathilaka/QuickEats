import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:provider/provider.dart';
import 'package:quickeats_customer/core/constants/app_colors.dart';
import 'package:quickeats_customer/viewmodels/cart_viewmodel.dart';
import 'package:quickeats_customer/viewmodels/order_viewmodel.dart';

class CheckoutScreen extends StatefulWidget {
  const CheckoutScreen({super.key});
  @override State<CheckoutScreen> createState() => _CheckoutScreenState();
}

class _CheckoutScreenState extends State<CheckoutScreen> {
  String _paymentMethod = 'wallet';

  @override Widget build(BuildContext context) {
    final cart = context.watch<CartViewModel>();
    final orderVM = context.watch<OrderViewModel>();
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

        // Payment method
        Container(padding: const EdgeInsets.all(16), decoration: BoxDecoration(color: AppColors.white, borderRadius: BorderRadius.circular(12), border: Border.all(color: AppColors.border)),
          child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
            const Text('Payment Method', style: TextStyle(fontSize: 16, fontWeight: FontWeight.w600)),
            const SizedBox(height: 12),
            RadioListTile<String>(value: 'wallet', groupValue: _paymentMethod, onChanged: (v) => setState(() => _paymentMethod = v!),
              title: const Text('Wallet'), secondary: const Icon(Icons.account_balance_wallet, color: AppColors.primary), contentPadding: EdgeInsets.zero),
            RadioListTile<String>(value: 'card', groupValue: _paymentMethod, onChanged: (v) => setState(() => _paymentMethod = v!),
              title: const Text('Card (PayHere)'), secondary: const Icon(Icons.credit_card, color: AppColors.secondary), contentPadding: EdgeInsets.zero),
          ]),
        ),
        const SizedBox(height: 24),

        if (orderVM.error != null) Padding(padding: const EdgeInsets.only(bottom: 12), child: Text(orderVM.error!, style: const TextStyle(color: AppColors.error))),

        SizedBox(height: 54, child: FilledButton(
          onPressed: orderVM.isPlacing || cart.isEmpty ? null : () async {
            final order = await orderVM.placeOrder(shopId: cart.shopId!, paymentMethod: _paymentMethod, items: cart.items);
            if (order != null && mounted) { cart.clear(); context.go('/orders/${order.id}'); }
          },
          style: FilledButton.styleFrom(backgroundColor: AppColors.primary, shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12))),
          child: orderVM.isPlacing
            ? const SizedBox(height: 22, width: 22, child: CircularProgressIndicator(strokeWidth: 2, color: Colors.white))
            : Text('Place Order — ${cart.formattedTotal}', style: const TextStyle(fontSize: 16, fontWeight: FontWeight.w600)),
        )),
      ])),
    );
  }
}
