import 'dart:async';
import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:provider/provider.dart';
import 'package:quickeats_shop/core/constants/app_colors.dart';
import 'package:quickeats_shop/core/constants/order_status.dart';
import 'package:quickeats_shop/models/order.dart';
import 'package:quickeats_shop/viewmodels/order_queue_viewmodel.dart';
import 'package:quickeats_shop/views/widgets/order_card.dart';

class OrderQueueScreen extends StatefulWidget {
  const OrderQueueScreen({super.key});
  @override
  State<OrderQueueScreen> createState() => _OrderQueueScreenState();
}

class _OrderQueueScreenState extends State<OrderQueueScreen> {
  late OrderQueueViewModel _vm;
  StreamSubscription<String>? _noticeSub;

  @override
  void initState() {
    super.initState();
    _vm = context.read<OrderQueueViewModel>();
    _noticeSub = _vm.notices.listen((text) {
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(SnackBar(
        content: Text(text),
        backgroundColor: Colors.red.shade700,
        behavior: SnackBarBehavior.floating,
      ));
    });
    // After the first frame: loading notifies listeners, which must not happen mid-build.
    WidgetsBinding.instance.addPostFrameCallback((_) { if (mounted) _vm.init(); });
  }

  @override
  void dispose() {
    _noticeSub?.cancel();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: AppColors.surface,
      appBar: AppBar(
        title: const Text('Order Queue', style: TextStyle(fontWeight: FontWeight.w700)),
        backgroundColor: AppColors.white,
        foregroundColor: AppColors.textPrimary,
        elevation: 0,
        actions: [
          IconButton(icon: const Icon(Icons.refresh), onPressed: () => _vm.loadOrders()),
        ],
      ),
      body: Consumer<OrderQueueViewModel>(
        builder: (context, vm, _) {
          if (vm.isLoading && vm.allOrders.isEmpty) {
            return const Center(child: CircularProgressIndicator());
          }
          if (vm.error != null && vm.allOrders.isEmpty) {
            return Center(
              child: Padding(
                padding: const EdgeInsets.all(24),
                child: Column(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    Icon(Icons.wifi_off, size: 56, color: AppColors.textMuted),
                    const SizedBox(height: 12),
                    Text("Couldn't load orders", style: TextStyle(fontSize: 16, color: AppColors.textSecondary)),
                    const SizedBox(height: 4),
                    Text('Check your internet connection and try again.', textAlign: TextAlign.center, style: TextStyle(fontSize: 13, color: AppColors.textMuted)),
                    const SizedBox(height: 16),
                    FilledButton.icon(onPressed: vm.loadOrders, icon: const Icon(Icons.refresh), label: const Text('Retry')),
                  ],
                ),
              ),
            );
          }
          return Column(
            children: [
              _buildStatusTabs(vm),
              _buildCountBar(vm),
              Expanded(child: _buildOrderList(vm)),
            ],
          );
        },
      ),
    );
  }

  Widget _buildStatusTabs(OrderQueueViewModel vm) {
    final tabs = [OrderStatus.confirmed, OrderStatus.accepted, OrderStatus.pickedUp];
    return Container(
      color: AppColors.white,
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
      child: SingleChildScrollView(
        scrollDirection: Axis.horizontal,
        child: Row(
          children: tabs.map((status) {
            final isSelected = vm.filterStatus == status;
            final count = vm.countFor(status);
            return Padding(
              padding: const EdgeInsets.only(right: 8),
              child: FilterChip(
                selected: isSelected,
                label: Text('${status.label} ($count)'),
                labelStyle: TextStyle(
                  fontSize: 13,
                  fontWeight: isSelected ? FontWeight.w600 : FontWeight.w400,
                  color: isSelected ? AppColors.white : AppColors.textSecondary,
                ),
                selectedColor: AppColors.forStatus(status.value),
                backgroundColor: AppColors.surfaceAlt,
                side: BorderSide.none,
                onSelected: (_) => vm.setFilter(status),
              ),
            );
          }).toList(),
        ),
      ),
    );
  }

  Widget _buildCountBar(OrderQueueViewModel vm) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 10),
      color: AppColors.white,
      child: Row(
        children: [
          _countChip('New', vm.countFor(OrderStatus.confirmed), AppColors.statusConfirmed),
          const SizedBox(width: 12),
          _countChip('In Progress', vm.countFor(OrderStatus.accepted), AppColors.statusPreparing),
        ],
      ),
    );
  }

  Widget _countChip(String label, int count, Color color) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
      decoration: BoxDecoration(color: color.withOpacity(0.1), borderRadius: BorderRadius.circular(20)),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Container(width: 8, height: 8, decoration: BoxDecoration(color: color, shape: BoxShape.circle)),
          const SizedBox(width: 6),
          Text('$count $label', style: TextStyle(fontSize: 12, fontWeight: FontWeight.w600, color: color)),
        ],
      ),
    );
  }

  Future<void> _advance(OrderQueueViewModel vm, Order order) async {
    final ok = await vm.updateStatus(order.id, order.status.nextShopAction!);
    if (!ok && mounted) {
      ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(vm.error ?? 'Failed to update the order')));
    }
  }

  Widget _buildOrderList(OrderQueueViewModel vm) {
    final orders = vm.orders;
    if (orders.isEmpty) {
      return Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(Icons.inbox_outlined, size: 64, color: AppColors.textMuted),
            const SizedBox(height: 12),
            Text('No ${vm.filterStatus.label.toLowerCase()} orders', style: TextStyle(fontSize: 16, color: AppColors.textSecondary)),
          ],
        ),
      );
    }

    return RefreshIndicator(
      onRefresh: vm.loadOrders,
      child: ListView.builder(
        padding: const EdgeInsets.all(12),
        itemCount: orders.length,
        itemBuilder: (context, index) {
          final order = orders[index];
          return OrderCard(
            order: order,
            onTap: () => context.push('/orders/${order.id}'),
            onAction: order.status.nextShopAction != null
              ? () => _advance(vm, order)
              : null,
          );
        },
      ),
    );
  }
}
