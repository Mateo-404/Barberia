package com.barber.barberBackend.config;

import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.context.annotation.Profile;
import org.springframework.web.cors.CorsConfiguration;
import org.springframework.web.cors.UrlBasedCorsConfigurationSource;
import org.springframework.web.filter.CorsFilter;

@Configuration
public class CorsConfig {

    @Bean
    @Profile("!prod")
    /*
     * Perfiles de desarrollo (default, h2, test, etc.): permiten cualquier
     * puerto en localhost para que Vite funcione sin importar el puerto que
     * asigne (5173, 5174, etc.). Nunca debe estar activo en producción.
     */
    public CorsFilter corsFilterDev() {
        CorsConfiguration config = new CorsConfiguration();
        config.addAllowedOriginPattern("http://localhost:*");
        config.addAllowedOriginPattern("http://127.0.0.1:*");
        config.addAllowedMethod("*");
        config.addAllowedHeader("*");
        config.setAllowCredentials(true);

        UrlBasedCorsConfigurationSource source = new UrlBasedCorsConfigurationSource();
        source.registerCorsConfiguration("/**", config);
        return new CorsFilter(source);
    }

    @Bean
    @Profile("prod")
    /*
     * Perfil de producción (Railway): solo permite el origen real de GitHub
     * Pages donde está deployado el frontend. Cualquier otro origen es
     * rechazado. Al agregar un nuevo frontend (ej: dominio custom), agregar
     * su origen acá.
     *
     * URL actual del frontend: https://mateo-404.github.io/Barberia/
     * Rama: frontend (ver README.md → Enlaces / GitHub Pages)
     */
    public CorsFilter corsFilterProd() {
        CorsConfiguration config = new CorsConfiguration();
        config.addAllowedOrigin("https://mateo-404.github.io");
        config.addAllowedMethod("*");
        config.addAllowedHeader("*");

        UrlBasedCorsConfigurationSource source = new UrlBasedCorsConfigurationSource();
        source.registerCorsConfiguration("/**", config);
        return new CorsFilter(source);
    }
}
