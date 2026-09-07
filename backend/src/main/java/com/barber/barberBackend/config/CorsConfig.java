package com.barber.barberBackend.config;

import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.context.annotation.Profile;
import java.util.List;

import org.springframework.web.cors.CorsConfiguration;
import org.springframework.web.cors.CorsConfigurationSource;
import org.springframework.web.cors.UrlBasedCorsConfigurationSource;

@Configuration
public class CorsConfig {

    /*
     * Nombre de bean explícito y compartido entre perfiles para poder
     * inyectarlo con @Qualifier desde SecurityConfig. Es necesario porque
     * Spring MVC ya registra otro CorsConfigurationSource
     * (mvcHandlerMappingIntrospector), lo que causa ambigüedad por tipo.
     */
    public static final String CORS_SOURCE_BEAN = "barberCorsConfigurationSource";

    @Bean(CORS_SOURCE_BEAN)
    @Profile("!prod")
    /*
     * Perfiles de desarrollo (default, h2, test, etc.): solo los puertos que
     * Vite realmente usa. Nunca debe estar activo en producción.
     */
    public CorsConfigurationSource corsConfigurationSourceDev() {
        CorsConfiguration config = new CorsConfiguration();
        config.setAllowedOriginPatterns(List.of(
                "http://localhost:5173",
                "http://localhost:5174",
                "http://127.0.0.1:5173",
                "http://127.0.0.1:5174"));
        config.setAllowedMethods(List.of("*"));
        config.setAllowedHeaders(List.of("*"));
        config.setAllowCredentials(true);

        UrlBasedCorsConfigurationSource source = new UrlBasedCorsConfigurationSource();
        source.registerCorsConfiguration("/**", config);
        return source;
    }

    @Bean(CORS_SOURCE_BEAN)
    @Profile("prod")
    /*
     * Perfil de producción (Railway): solo el origen real de GitHub Pages.
     * URL actual del frontend: https://mateo-404.github.io/Barberia/
     */
    public CorsConfigurationSource corsConfigurationSourceProd() {
        CorsConfiguration config = new CorsConfiguration();
        config.addAllowedOrigin("https://mateo-404.github.io");
        config.addAllowedMethod("*");
        config.addAllowedHeader("*");

        UrlBasedCorsConfigurationSource source = new UrlBasedCorsConfigurationSource();
        source.registerCorsConfiguration("/**", config);
        return source;
    }
}
