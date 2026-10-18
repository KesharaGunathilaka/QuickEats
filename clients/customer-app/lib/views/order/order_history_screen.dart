import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:provider/provider.dart';
import 'package:intl/intl.dart';
import 'package:quickeats_customer/core/constants/app_colors.dart';
import 'package:quickeats_customer/viewmodels/order_viewmodel.dart';

class OrderHistoryScreen extends StatefulWidget {
  const OrderHistoryScreen({super.key});
  @override State<OrderHistoryScreen> createState() => _OrderHistoryScreenState();
}

class _OrderHistoryScreenState extends State<OrderHistoryScreen> {
  @override void initState() { super.initState(); context.read<OrderViewModel>().loadOrders(); }

  @override Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('My Orders', style: TextStyle(fontWeight: FontWeight.w700)), backgroundColor: AppColors.white, elevation: 0),
      body: Consumer<OrderViewModel>(builder: (_, vm, __) {
        if (vm.isLoading) return const Center(child: CircularProgressIndicator());
        if (vm.orders.isEmpty) return const Center(child: Text('No orders yet'));
        return RefreshIndicator(onRefresh: vm.loadOrders, child: ListView.builder(
          padding: const EdgeInsets.all(16), itemCount: vm.orders.length,
          itemBuilder: (_, i) {
            final order = vm.orders[i];
            final color = order.status.color;
            return Card(margin: const EdgeInsets.only(bottom: 10), elevation: 0,
              shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12), side: const BorderSide(color: AppColors.border)),
              child: InkWell(onTap: () => context.push('/orders/${order.id}'), borderRadius: BorderRadius.circular(12),
                child: Padding(padding: const EdgeInsets.all(14), child: Row(children: [
                  Container(width: 44, height: 44, decoration: BoxDecoration(color: color.withOpacity(0.1), borderRadius: BorderRadius.circular(10)),
                    child: Icon(order.status.icon, color: color, size: 22)),
                  const SizedBox(width: 14),
                  Expanded(child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
                    Text('#${order.shortId}', style: const TextStyle(fontWeight: FontWeight.w600)),
                    Text('${order.items.length} items · ${order.formattedTotal}', style: const TextStyle(fontSize: 13, color: AppColors.textSecondary)),
                    Text(DateFormat('MMM d, h:mm a').format(order.createdAt), style: const TextStyle(fontSize: 12, color: AppColors.textMuted)),
                  ])),
                  Container(padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4), decoration: BoxDecoration(color: color.withOpacity(0.1), borderRadius: BorderRadius.circular(8)),
                    child: Text(order.status.label, style: TextStyle(fontSize: 11, fontWeight: FontWeight.w600, color: color))),
                ])),
              ),
            );
          },
        ));
      }),
    );
  }
}
