package controller

/*func (r *CorePodReconciler) deleteExternalResources(ctx context.Context, corepod *webappv1.CorePod, l logr.Logger) error {
	l.Info("[COREPOD]: Entered Delete ")
	// Delete CORE_EXT
	corext := &webappv1.CoreExt{}
	err := r.Client.Get(ctx, types.NamespacedName{Name: corepod.Name + "-ext", Namespace: corepod.Namespace}, corext)

	if err != nil {
		return err
	}
	r.Client.Update(ctx, corext, &client.UpdateOptions{})

	return nil
} */
