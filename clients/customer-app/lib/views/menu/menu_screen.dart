import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:provider/provider.dart';
import 'package:quickeats_customer/core/constants/app_colors.dart';
import 'package:quickeats_customer/viewmodels/shops_viewmodel.dart';
import 'package:quickeats_customer/viewmodels/cart_viewmodel.dart';

class MenuScreen extends StatefulWidget {
  final String shopId;
  const MenuScreen({super.key, required this.shopId});
  @override State<MenuScreen> createState() => _MenuScreenState();
}

class _MenuScreenState extends State<MenuScreen> {
  @override void initState() { super.initState(); context.read<ShopsViewModel>().selectShop(widget.shopId); }

  @override Widget build(BuildContext context) {
    final cart = context.watch<CartViewModel>();
    return Consumer<ShopsViewModel>(builder: (_, vm, __) {
      return Scaffold(
        appBar: AppBar(title: Text(vm.selectedShop?.name ?? 'Menu', style: const TextStyle(fontWeight: FontWeight.w700)), backgroundColor: AppColors.white, elevation: 0),
        body: vm.isLoading ? const Center(child: CircularProgressIndicator()) : ListView(
          padding: const EdgeInsets.all(16),
          children: vm.menuByCategory.entries.map((entry) => Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Padding(padding: const EdgeInsets.symmetric(vertical: 8), child: Text(entry.key, style: const TextStyle(fontSize: 16, fontWeight: FontWeight.w600))),
              ...entry.value.map((item) => Card(
                margin: const EdgeInsets.only(bottom: 8), elevation: 0,
                shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(10), side: const BorderSide(color: AppColors.border)),
                child: Padding(padding: const EdgeInsets.all(14), child: Row(children: [
                  Expanded(child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
                    Text(item.name, style: const TextStyle(fontSize: 15, fontWeight: FontWeight.w600)),
                    if (item.description.isNotEmpty) Text(item.description, style: const TextStyle(fontSize: 12, color: AppColors.textSecondary), maxLines: 2),
                    const SizedBox(height: 4),
                    Text(item.formattedPrice, style: const TextStyle(fontSize: 14, fontWeight: FontWeight.w600, color: AppColors.primary)),
                  ])),
                  const SizedBox(width: 12),
                  item.available && item.stock > 0
                    ? SizedBox(height: 36, child: FilledButton.icon(
                        onPressed: () { cart.addItem(item, shopId: widget.shopId, shopName: vm.selectedShop?.name); ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text('${item.name} added to cart'), duration: const Duration(seconds: 1))); },
                        icon: const Icon(Icons.add, size: 18), label: const Text('Add'),
                        style: FilledButton.styleFrom(backgroundColor: AppColors.primary, padding: const EdgeInsets.symmetric(horizontal: 12), shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(8))),
                      ))
                    : Container(padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8), decoration: BoxDecoration(color: AppColors.surfaceAlt, borderRadius: BorderRadius.circular(8)),
                        child: const Text('Sold Out', style: TextStyle(fontSize: 12, color: AppColors.textMuted, fontWeight: FontWeight.w500))),
                ])),
              )),
            ],
          )).toList(),
        ),
        bottomNavigationBar: cart.isEmpty ? null : Container(
          padding: const EdgeInsets.all(16), decoration: const BoxDecoration(color: AppColors.white, border: Border(top: BorderSide(color: AppColors.border))),
          child: SafeArea(child: SizedBox(height: 50, child: FilledButton(
            onPressed: () => context.go('/cart'),
            style: FilledButton.styleFrom(backgroundColor: AppColors.primary, shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12))),
            child: Text('View Cart (${cart.itemCount} items) — ${cart.formattedTotal}', style: const TextStyle(fontSize: 15, fontWeight: FontWeight.w600)),
          ))),
        ),
      );
    });
  }
}
