import 'package:flutter/material.dart';

class AppColors {
  static const primary = Color(0xFF2563EB);
  static const primaryDark = Color(0xFF1D4ED8);
  static const secondary = Color(0xFF059669);
  static const error = Color(0xFFDC2626);
  static const warning = Color(0xFFF59E0B);
  static const surface = Color(0xFFF8FAFC);
  static const surfaceAlt = Color(0xFFF1F5F9);
  static const border = Color(0xFFE2E8F0);
  static const textPrimary = Color(0xFF1E293B);
  static const textSecondary = Color(0xFF64748B);
  static const textMuted = Color(0xFF94A3B8);
  static const white = Colors.white;

  static const statusConfirmed = Color(0xFF3B82F6);
  static const statusPreparing = Color(0xFFF97316);
  static const statusReady = Color(0xFF22C55E);

  static Color forStatus(String status) {
    switch (status) {
      case 'PENDING_PAYMENT': return const Color(0xFFF59E0B);
      case 'CONFIRMED': return const Color(0xFF3B82F6);
      case 'ACCEPTED': return const Color(0xFF8B5CF6);
      case 'PREPARING': return const Color(0xFFF97316);
      case 'READY': return const Color(0xFF22C55E);
      case 'PICKED_UP': return const Color(0xFF6B7280);
      case 'CANCELLED': case 'REJECTED': return const Color(0xFFEF4444);
      default: return const Color(0xFF6B7280);
    }
  }
}
