
export const requirePermission = (permissions: any, role: string): boolean => {
    return permissions && permissions.role && permissions.role === role;
}