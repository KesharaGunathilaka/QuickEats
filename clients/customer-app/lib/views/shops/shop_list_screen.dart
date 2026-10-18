import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:provider/provider.dart';
import 'package:quickeats_customer/core/constants/app_colors.dart';
import 'package:quickeats_customer/viewmodels/auth_viewmodel.dart';
import 'package:quickeats_customer/viewmodels/cart_viewmodel.dart';
import 'package:quickeats_customer/viewmodels/shops_viewmodel.dart';

class ShopListScreen extends StatefulWidget {
  const ShopListScreen({super.key});
  @override State<ShopListScreen> createState() => _ShopListScreenState();
}

class _ShopListScreenState extends State<ShopListScreen> {
  @override void initState() { super.initState(); context.read<ShopsViewModel>().loadShops(); }

  Future<void> _logout() async {
    final ok = await showDialog<bool>(context: context, builder: (c) => AlertDialog(
      title: const Text('Log out?'),
      actions: [TextButton(onPressed: () => Navigator.pop(c, false), child: const Text('Cancel')),
                FilledButton(onPressed: () => Navigator.pop(c, true), child: const Text('Log out'))],
    ));
    if (ok != true || !mounted) return;
    context.read<CartViewModel>().clear();
    await context.read<AuthViewModel>().logout();
    if (mounted) context.go('/login');
  }

  @override Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Shops', style: TextStyle(fontWeight: FontWeight.w700)), backgroundColor: AppColors.white, foregroundColor: AppColors.textPrimary, elevation: 0,
        actions: [IconButton(icon: const Icon(Icons.logout), tooltip: 'Log out', onPressed: _logout)]),
      body: Consumer<ShopsViewModel>(builder: (_, vm, __) {
        if (vm.isLoading) return const Center(child: CircularProgressIndicator());
        if (vm.shops.isEmpty) return const Center(child: Text('No shops available'));
        return RefreshIndicator(
          onRefresh: vm.loadShops,
          child: ListView.builder(
            padding: const EdgeInsets.all(16), itemCount: vm.shops.length,
            itemBuilder: (_, i) {
              final shop = vm.shops[i];
              return Card(
                margin: const EdgeInsets.only(bottom: 12), elevation: 0,
                shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12), side: const BorderSide(color: AppColors.border)),
                child: InkWell(
                  onTap: () => context.push('/shops/${shop.id}/menu'),
                  borderRadius: BorderRadius.circular(12),
                  child: Padding(padding: const EdgeInsets.all(16), child: Row(children: [
                    Container(width: 56, height: 56, decoration: BoxDecoration(color: AppColors.primaryLight, borderRadius: BorderRadius.circular(12)),
                      child: const Icon(Icons.storefront, color: AppColors.primary, size: 28)),
                    const SizedBox(width: 14),
                    Expanded(child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
                      Text(shop.name, style: const TextStyle(fontSize: 16, fontWeight: FontWeight.w600)),
                      const SizedBox(height: 2),
                      Text(shop.description, style: const TextStyle(fontSize: 13, color: AppColors.textSecondary), maxLines: 1, overflow: TextOverflow.ellipsis),
                    ])),
                    Container(
                      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
                      decoration: BoxDecoration(color: shop.isOpen ? AppColors.secondary.withOpacity(0.1) : AppColors.surfaceAlt, borderRadius: BorderRadius.circular(12)),
                      child: Text(shop.isOpen ? 'Open' : 'Closed', style: TextStyle(fontSize: 12, fontWeight: FontWeight.w600, color: shop.isOpen ? AppColors.secondary : AppColors.textMuted)),
                    ),
                  ])),
                ),
              );
            },
          ),
        );
      }),
    );
  }
}
